import argparse
import json
import threading

from flask import Flask, request

from learnNetwork import Network


def parse_args():
    parser = argparse.ArgumentParser(description='Please provide mysql arguments')
    parser.add_argument("--mysql_host", dest="mysql_host", required=True, help="Provide mysql DB host")
    parser.add_argument("--mysql_user", dest="mysql_user", required=True, help="Provide mysql username")
    parser.add_argument("--mysql_pass", dest="mysql_pass", required=True, help="Provide mysql password")
    parser.add_argument("--mysql_db", dest="mysql_db", default="ob", help="Provide mysql database name")
    parser.add_argument("--port", dest='port', default=8181, help="Provide server http port")
    parser.add_argument("--manual", dest="manual_classification", action="store_true")
    return parser.parse_args()


def server(args, net):
    api = Flask(__name__)

    @api.route('/checkListing', methods=['POST'])
    def process_listing():
        return json.dumps(net.test_listing(request.json)[0])

    @api.route('/checkListings', methods=['POST'])
    def process_listings():
        return json.dumps(net.test_listing(request.json))

    api.run(port=args.port)


def learning_network(args, net):
    return net.start_for_manual_user_classification(args)


def main():
    args = parse_args()

    net = Network(args)
    if args.manual_classification:
        server_thread = threading.Thread(target=server, args=(args, net))
        ml_thread = threading.Thread(target=learning_network, args=(args, net))

        server_thread.start()
        ml_thread.start()

        server_thread.join()
        ml_thread.join()
    else:
        server(args, net)


if __name__ == '__main__':
    main()
