import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  vus: 10,
  duration: '10s',
};

export default function () {
  let res = http.post('http://localhost:8080/api/v1/chats/query', JSON.stringify({
    sessionId: "",
    query: "What is contract law?",
    language: "en"
  }), { headers: { 'Content-Type': 'application/json' } });
  check(res, { 'status was 200': (r) => r.status == 200 });
  sleep(1);
}