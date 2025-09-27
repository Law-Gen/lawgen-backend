import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  vus: 10,
  duration: '10s',
};

const BASE = 'http://localhost:8080/api/v1/quizzes';

export default function () {
  // 1. Fetch all categories
  let res1 = http.get(`${BASE}/categories`);
  check(res1, { 'GET /categories status was 200 or 404': (r) => r.status == 200 || r.status == 404});

  // // 2. Fetch a specific quiz (replace 'sampleQuizId' with a real one)
  // let res3 = http.get(`${BASE}/sampleQuizId`);
  // check(res3, { 'GET /:quizId status was 200 or 404': (r) => r.status == 200 || r.status == 404 });

  // // 3. Fetch questions for a quiz (replace 'sampleQuizId' with a real one)
  // let res4 = http.get(`${BASE}/sampleQuizId/questions`);
  // check(res4, { 'GET /:quizId/questions status was 200 or 404': (r) => r.status == 200 || r.status == 404 });

  sleep(1);
}