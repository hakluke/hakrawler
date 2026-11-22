#!/usr/bin/env python3
"""Tiny local web server for hakrawler benchmarking.

Serves pages with a configurable number of links, counts HTTP requests and
TCP connections, and exposes /__stats and /__reset for measurement.

Usage:
    python3 bench/server.py
    # env: BENCH_PORT (8899), BENCH_LINKS (20), BENCH_DELAY (0.03s), BENCH_PAGES (60)

    curl -s http://127.0.0.1:8899/__reset
    seq 0 99 | sed 's|^|http://127.0.0.1:8899/p/|' | ./hakrawler -d 1 -t 8 > /dev/null
    curl -s http://127.0.0.1:8899/__stats
"""
import json
import socket
import os
import sys
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

LINKS = int(os.environ.get("BENCH_LINKS", "20"))
DELAY = float(os.environ.get("BENCH_DELAY", "0"))  # seconds per response
PORT = int(os.environ.get("BENCH_PORT", "8899"))
PAGES = int(os.environ.get("BENCH_PAGES", "60"))


def page(title, links):
    items = "".join('<a href="/p/%d">link %d</a> ' % (i, i) for i in links)
    return (
        "<!doctype html><html><head><title>%s</title>"
        '<script src="/static/app.js"></script></head><body>'
        "%s<form action=\"/search\"><input name=q></form></body></html>"
    ) % (title, items)


class Handler(BaseHTTPRequestHandler):
    protocol_version = "HTTP/1.1"

    def setup(self):
        super().setup()
        try:
            self.connection.setsockopt(socket.IPPROTO_TCP, socket.TCP_NODELAY, 1)
        except OSError:
            pass
        self.server.connections += 1

    def log_message(self, *args):
        pass

    def _send(self, body, ctype="text/html; charset=utf-8"):
        if DELAY:
            import time
            time.sleep(DELAY)
        raw = body.encode()
        self.send_response(200)
        self.send_header("Content-Type", ctype)
        self.send_header("Content-Length", str(len(raw)))
        self.end_headers()
        self.wfile.write(raw)
        self.server.requests += 1

    def do_GET(self):
        path = self.path.split("?", 1)[0]
        if path == "/__stats":
            raw = json.dumps({
                "requests": self.server.requests,
                "connections": self.server.connections,
            }).encode()
            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.send_header("Content-Length", str(len(raw)))
            self.end_headers()
            self.wfile.write(raw)
            return
        if path == "/__reset":
            self.server.requests = 0
            self.server.connections = 0
            raw = b"ok"
            self.send_response(200)
            self.send_header("Content-Length", str(len(raw)))
            self.end_headers()
            self.wfile.write(raw)
            return
        if path == "/robots.txt":
            self._send("User-agent: *\nAllow: /\n", "text/plain")
            return
        if path == "/static/app.js":
            self._send("// js\n", "application/javascript")
            return
        if path == "/":
            self._send(page("index", range(min(LINKS, PAGES))))
            return
        if path.startswith("/p/"):
            try:
                n = int(path[3:])
            except ValueError:
                n = 0
            links = [(n + 1) % PAGES, (n + 2) % PAGES, (n + 7) % PAGES]
            self._send(page("page %d" % n, links))
            return
        self._send(page("other", []))


if __name__ == "__main__":
    srv = ThreadingHTTPServer(("127.0.0.1", PORT), Handler)
    srv.requests = 0
    srv.connections = 0
    print("serving on %d (links=%d delay=%s pages=%d)" % (PORT, LINKS, DELAY, PAGES), flush=True)
    srv.serve_forever()
