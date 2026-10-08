"""mock OpenAI 兼容服务（验证 aiops LLM 网关）：POST /v1/chat/completions 返回固定口径结论"""
import json
from http.server import BaseHTTPRequestHandler, HTTPServer


class H(BaseHTTPRequestHandler):
    def do_POST(self):
        n = int(self.headers.get('content-length', 0))
        body = json.loads(self.rfile.read(n) or b'{}')
        prompt = ''
        for m in body.get('messages', []):
            prompt += m.get('content', '')
        if 'ImagePullBackOff' in prompt:
            verdict = '根因：镜像拉取失败（快照含 ImagePullBackOff 证据）'
        elif len(prompt) > 100:
            verdict = '根因：待定（快照 %d 字符）' % len(prompt)
        else:
            verdict = '根因：短提问 %d 字符' % len(prompt)
        out = {
            'id': 'mock-1', 'object': 'chat.completion', 'model': body.get('model', 'mock'),
            'choices': [{'index': 0,
                         'message': {'role': 'assistant',
                                     'content': '【MOCK】' + verdict + '。建议：核对镜像名/仓库凭据后重建。'},
                         'finish_reason': 'stop'}],
            'usage': {'prompt_tokens': len(prompt), 'completion_tokens': 20,
                      'total_tokens': len(prompt) + 20},
        }
        resp = json.dumps(out).encode()
        self.send_response(200)
        self.send_header('Content-Type', 'application/json')
        self.send_header('Content-Length', str(len(resp)))
        self.end_headers()
        self.wfile.write(resp)

    def log_message(self, fmt, *args):
        pass


if __name__ == '__main__':
    HTTPServer(('127.0.0.1', 8890), H).serve_forever()
