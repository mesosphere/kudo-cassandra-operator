#!/usr/bin/env python3

from http.server import BaseHTTPRequestHandler, HTTPServer
from medusa.medusacli import cli
import sys
import subprocess

PORT_NUMBER = 8082

# This class will handles any incoming request from
# the browser
class myHandler(BaseHTTPRequestHandler):

    # Handler for the GET requests
    def do_GET(self):
        if self.path == "/backup":
            self.doBackup()
            return

        self.send_response(200)
        self.send_header('Content-type', 'text/html')
        self.end_headers()
        # Send the html message
        msg = "Unknown command!(" + self.path + ")"
        self.wfile.write(msg.encode("UTF-8"))
        return

    def doBackup(self):
        self.send_response(200)
        self.send_header('Content-type', 'text/html')
        self.end_headers()
        # Send the html message
        self.wfile.write("Starting medusa\n".encode("UTF-8"))

        res = subprocess.run( ["python3", "/usr/local/bin/medusa", "backup"], stdout=subprocess.PIPE, stderr=self.wfile)

        end_message = "Done: " + str(res.returncode) + "\n"

        self.wfile.write(end_message.encode("UTF-8"))

        return


try:
    # Create a web server and define the handler to manage the
    # incoming request
    server = HTTPServer(('', PORT_NUMBER), myHandler)
    print('Started httpserver on port ', PORT_NUMBER)

    # Wait forever for incoming htto requests
    server.serve_forever()

except KeyboardInterrupt:
    print('^C received, shutting down the web server')
    server.socket.close()
