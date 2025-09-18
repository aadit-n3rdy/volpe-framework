FROM python:3-slim

RUN pip install grpcio grpcio-tools

WORKDIR /home/
COPY . .
RUN python -m grpc_tools.protoc -I. --python_out=. --grpc_python_out=. ./greeter.proto

CMD [ "bash" ]