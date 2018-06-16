import sys
import errno

sys.path.append('gen-py')
from server import server
from server.ttypes import *

from thrift import Thrift
from thrift.transport import TSocket
from thrift.transport import TTransport
from thrift.protocol import TBinaryProtocol

def connect():
  transport = TSocket.TSocket('127.0.0.1', 9090)
  transport = TTransport.TBufferedTransport(transport)
  protocol = TBinaryProtocol.TBinaryProtocol(transport)
  client = server.Client(protocol)
  transport.open()
  return client

def main():
  client = connect()
  client.set_fault(['write'], False, 0, 0, "", False, 50000000, False)
