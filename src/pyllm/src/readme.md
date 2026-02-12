# llyb-PyLLM

## Project Structure
```plaintext

├── Dockerfile                 # Docker configuration for building the app
├── requirements.txt           # Python dependencies
├── readme.md
├── src/
│   ├── config/                # Configuration files
│   │   └── settings.py        # Configuration settings, such as API keys and model
│   ├── grpc/                  # gRPC related files
│   │   ├── grpc_server.py     # Python implementation of the gRPC server
│   │   ├── pyllm_pb2_grpc.py  # Generated gRPC client and server classes
│   │   ├── pyllm_pb2.py       # gRPC message definitions (generated from proto)
│   │   └── pyllm.proto        # gRPC service definitions
│   ├── prompts/               # Prompt templates
│   │   ├── load_prompts.py    # Loading prompt templates
│   │   └── prompts.json       # Prompt templates instance
│   ├── utils/                 # Utility functions
│   │   └── serialize.py       # Serialization function
│   └── llm.py                 # LangChain logic and model integration

```

## Installation Environment
```plaintext
pip install -r requirements.txt
```

## Generate gRPC code
```plaintext
python -m grpc_tools.protoc -I=./src/grpc --python_out=./src/grpc --grpc_python_out=./src/grpc ./src/grpc/pyllm.proto
```

## Running gRPC server
```plaintext
python src/grpc/grpc_server.py
```

## Future Development
>- RAG
>- Optimize Prompt Template