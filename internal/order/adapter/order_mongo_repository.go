package adapter

import (
	"context"
	"fmt"
	"time"

	_ "github.com/furutachiKurea/gorder/common/config"
	domain "github.com/furutachiKurea/gorder/order/domain/order"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	dbName   = viper.GetString("mongo.db-name")
	collName = viper.GetString("mongo.coll-name")
)

type OrderRepositoryMongo struct {
	db *mongo.Client
}

func NewOrderRepositoryMongo(db *mongo.Client) *OrderRepositoryMongo {
	return &OrderRepositoryMongo{db: db}
}

func (r *OrderRepositoryMongo) Create(ctx context.Context, order *domain.Order) (created *domain.Order, err error) {
	defer r.logWithTag("create", err, created)

	write := r.domainToMongo(order)
	res, err := r.collection().InsertOne(ctx, write)
	if err != nil {
		return nil, err
	}

	created = order
	created.ID = res.InsertedID.(primitive.ObjectID).Hex()
	return
}

func (r *OrderRepositoryMongo) Get(ctx context.Context, orderID, customerID string) (got *domain.Order, err error) {
	defer r.logWithTag("get", err, got)

	read := &orderModel{}
	mongoID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		err = fmt.Errorf("generate mongo id from orderID: %w", err)
		return
	}

	cond := bson.M{
		"_id": mongoID,
	}

	if err = r.collection().FindOne(ctx, cond).Decode(read); err != nil {
		return
	}

	if read == nil {
		return nil, domain.NotFoundError{OrderID: orderID}
	}

	return r.unmarshal(read), nil
}

// Update 先查找对应的 Order，然后 apply updateFn，再写入 Mongo
func (r *OrderRepositoryMongo) Update(ctx context.Context, updates *domain.Order) (err error) {
	defer r.logWithTag("update", err, nil)

	if updates == nil {
		panic("got nil order")
	}

	session, err := r.db.StartSession()
	if err != nil {
		return
	}
	defer session.EndSession(ctx)

	if err = session.StartTransaction(); err != nil {
		return
	}
	defer func() {
		if err == nil {
			_ = session.CommitTransaction(ctx)
		} else {
			_ = session.AbortTransaction(ctx)
		}
	}()

	// transaction in (end at defer)
	oldOrder, err := r.Get(ctx, updates.ID, updates.CustomerID)
	if err != nil {
		return
	}

	updated := r.updateOrder(oldOrder, updates)
	log.Debug().Any("order_update_to", updated).Msg("")

	mongoID, _ := primitive.ObjectIDFromHex(oldOrder.ID)
	res, err := r.collection().UpdateOne(
		ctx,
		bson.M{"_id": mongoID},
		bson.M{"$set": bson.M{
			"id":           mongoID,
			"status":       updated.Status,
			"payment_link": updated.PaymentLink,
		}},
	)
	if err != nil {
		return
	}

	r.logWithTag("finish_update", err, res)
	return
}

// collection 获取订单 collection
func (r *OrderRepositoryMongo) collection() *mongo.Collection {
	return r.db.Database(dbName).Collection(collName)
}

func (r *OrderRepositoryMongo) domainToMongo(order *domain.Order) *orderModel {
	return &orderModel{
		MongoID:     primitive.NewObjectID(),
		ID:          order.ID,
		CustomerID:  order.CustomerID,
		Status:      order.Status,
		PaymentLink: order.PaymentLink,
		Items:       order.Items,
	}
}

func (r *OrderRepositoryMongo) unmarshal(m *orderModel) *domain.Order {
	return &domain.Order{
		ID:          m.MongoID.Hex(),
		CustomerID:  m.CustomerID,
		Status:      m.Status,
		PaymentLink: m.PaymentLink,
		Items:       m.Items,
	}
}

func (r *OrderRepositoryMongo) logWithTag(tag string, err error, result any) {
	l := log.With().
		Str("repository", "order_repository_mongo").
		Time("performed_at", time.Now()).
		Err(err).
		Any("result", result).
		Logger()

	if err != nil {
		l.Error().Msgf("%s_failed", tag)
	} else {
		l.Info().Msgf("%s_succeeded", tag)
	}
}

// updateOrder 根据 old order 和 updates 生成新的 order
//
// PaymentLink 始终使用 updates 的值，使得在支付完成之后 PaymentLink 会被置空
func (r *OrderRepositoryMongo) updateOrder(old *domain.Order, updates *domain.Order) *domain.Order {
	res := &domain.Order{
		ID:          old.ID,
		CustomerID:  old.CustomerID,
		Status:      old.Status,
		PaymentLink: updates.PaymentLink,
		Items:       old.Items,
	}

	if updates.Status != "" && updates.Status != old.Status {
		res.UpdatesStatus(updates.Status)
	}

	return res
}

// orderModel MongoDB 的订单模型
type orderModel struct {
	MongoID     primitive.ObjectID `bson:"_id"`
	ID          string             `bson:"id"` // ID 与 MongoID 对应
	CustomerID  string             `bson:"customer_id"`
	Status      string             `bson:"status"`
	PaymentLink string             `bson:"payment_link"`
	Items       []*domain.Item     `bson:"items"`
}
