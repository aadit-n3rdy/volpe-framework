
import volpe_pb2_grpc as vp
import volpe_pb2 as pb
import grpc
import concurrent.futures

class VolpeGreeterServicer(vp.VolpeContainerServicer):
    def SayHello(self, request: pb.HelloRequest, context):
        return pb.HelloReply(message="hello " + request.name)
    
server = grpc.server(concurrent.futures.ThreadPoolExecutor(max_workers=10))
vp.add_VolpeContainerServicer_to_server(VolpeGreeterServicer(), server)
server.add_insecure_port("0.0.0.0:8081")
server.start()
server.wait_for_termination()
