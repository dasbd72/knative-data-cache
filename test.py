from http.server import BaseHTTPRequestHandler, HTTPServer

class SimpleHTTPRequestHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        # Print the requested URL
        client_ip, client_port = self.client_address
        print("Client IP:", client_ip)
        print("Client Port:", client_port)

        self.send_response(200)
        self.send_header('Content-type', 'text/html')
        self.end_headers()
        response = "Hello, this is a simple HTTP server response!"
        self.wfile.write(response.encode('utf-8'))

def run_server(port=8000):
    server_address = ('', port)
    httpd = HTTPServer(server_address, SimpleHTTPRequestHandler)
    print(f"Server started on port {port}")
    httpd.serve_forever()

if __name__ == '__main__':
    run_server()
