import os
from dotenv import load_dotenv

load_dotenv()

class Settings:
  # LLM
  # LLM_BASE_URL = os.getenv("LLM_BASE_URL")
  # LLM_API_KEY = os.getenv("LLM_API_KEY")
  LLM_BASE_URL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
  LLM_API_KEY = "sk-633db6243a3e4d72b3299ec564ff7845"
  MODEL = "deepseek-v3.1"

  # Milvus
  # MILVUS_HOST = os.getenv("MILVUS_HOST", "localhost")
  # MILVUS_PORT = int(os.getenv("MILVUS_PORT", 19530))

config = Settings()
