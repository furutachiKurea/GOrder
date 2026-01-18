import http from 'k6/http';
import {check, sleep} from 'k6';
import {Counter} from 'k6/metrics';

export const options = {
  // 降低并发用户数，减少数据库连接压力
  stages: [
    { duration: '20s', target: 100 },  // 从 200 降到 100
    { duration: '40s', target: 100 },
    { duration: '10s', target: 0 },
  ],
};

const successCounter = new Counter('successful_orders');
const failCounter = new Counter('failed_orders');

export default function () {
  const customerId = 'kp_testing';
  const itemId = 'item_hot';
  const quantity = 1;

  const url = `http://localhost:8082/api/customer/${customerId}/orders`;
  const payload = JSON.stringify({
    customer_id: customerId,
    items: [
      {
        id: itemId,
        quantity: quantity,
      },
    ],
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  const res = http.post(url, payload, params);

  const isSuccess = check(res, {
    'status is 200': (r) => r.status === 200,
  });

  if (isSuccess) {
    successCounter.add(1);
  } else {
    failCounter.add(1);
    // Optional: Log errors to understand why it failed (lock, stock, etc.)
    // console.log(`Failed: ${res.status} ${res.body}`);
  }

  // 增加请求间隔，进一步降低 QPS
  sleep(0.2);  // 从 0.1 改为 0.2 秒
}
