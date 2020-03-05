#!/usr/bin/env python3

from http.server import BaseHTTPRequestHandler, HTTPServer
from urllib.parse import urlparse, parse_qs
from medusa.medusacli import cli
import sys
import subprocess

PORT_NUMBER = 8082


# This class will handles any incoming request from
# the browser
class myHandler(BaseHTTPRequestHandler):

    # Handler for the GET requests
    def do_GET(self):
        request_url = urlparse(self.path)

        if request_url.path == "/backup":
            self.doBackup(request_url)
            return

        self.send_response(200)
        self.send_header('Content-type', 'text/html')
        self.end_headers()
        # Send the html message
        msg = "Unknown command!(" + self.path + ")"
        self.wfile.write(msg.encode("UTF-8"))
        return

    def doBackup(self, request_url):
        self.send_response(200)
        self.send_header('Content-type', 'text/html')
        self.end_headers()
        # Send the html message

        cmd = ["python3", "/usr/local/bin/medusa"]

        query_params = parse_qs(request_url.query)

        # Global options first
        if 'fqdn' in query_params:
            cmd.extend(["--fqdn", query_params['fqdn'][0]])

        cmd.append("backup")

        # Backup args after the command
        if 'name' in query_params:
            cmd.extend(["--backup-name", query_params['name'][0]])

        self.wfile.write(("Starting medusa: " + ' '.join([str(v) for v in cmd]) + "\n").encode("UTF-8"))

        res = subprocess.run(args=cmd, stdout=subprocess.PIPE, stderr=self.wfile)

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
