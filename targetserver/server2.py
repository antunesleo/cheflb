from http.server import BaseHTTPRequestHandler, HTTPServer
import random
import time
import urllib.parse

class PrintMeHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        self.print_request_details()

    def do_POST(self):
        self.print_request_details()
        
    def do_PUT(self):
        self.print_request_details()

    def print_request_details(self):
        time.sleep(random.randint(1, 3))
        # Parse the full URL with query strings
        parsed_path = urllib.parse.urlparse(self.path)
        query_string = urllib.parse.parse_qs(parsed_path.query)

        # Print HTTP method
        print(f"HTTP Method: {self.command}")

        # Print headers
        print("Headers:")
        for key, value in self.headers.items():
            print(f"{key}: {value}")

        # Print body
        content_length = int(self.headers.get('Content-Length', 0))
        body = self.rfile.read(content_length).decode('utf-8')
        print(f"Body: {body}")

        # Print full URL with query strings
        full_url = f"{self.path}"
        print(f"Full URL: {full_url}")

        # Send response
        self.send_response(200)
        self.send_header("some", "header")
        self.flush_headers()
        self.end_headers()
        self.wfile.write(b"Request details printed to the console.")

def run(server_class=HTTPServer, handler_class=PrintMeHandler, port=8181):
    server_address = ('', port)
    httpd = server_class(server_address, handler_class)
    print(f"Starting server on port {port}")
    httpd.serve_forever()

if __name__ == '__main__':
    run()
