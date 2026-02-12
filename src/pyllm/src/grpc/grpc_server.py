import grpc
from concurrent import futures
import pyllm_pb2
import pyllm_pb2_grpc

from llm import ChatLLM
from config.settings import config
from utils.serialize import serialize_chunk_langchain

class pyLLMService(pyllm_pb2_grpc.pyLLMServiceServicer):
  def __init__(self):
    self.llm = ChatLLM(model=config.MODEL)

  def GenerateResponse(self, request, context):
    birth_date = request.birth_date
    birth_time = request.birth_time

    # 生成流式响应
    for chunk in self.llm.chain.stream({"BIRTH_DATE": birth_date, "BIRTH_TIME": birth_time}):
      # 将流式数据返回给客户端
      yield pyllm_pb2.LLMResponse(chunk=chunk.content)

def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    pyllm_pb2_grpc.add_pyLLMServiceServicer_to_server(pyLLMService(), server)
    server.add_insecure_port('[::]:50051')  # 启动服务监听端口
    server.start()
    print("gRPC server running on port 50051...")
    server.wait_for_termination()

if __name__ == '__main__':
    serve()

