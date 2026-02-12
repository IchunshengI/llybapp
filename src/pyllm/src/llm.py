from langchain_core.output_parsers import StrOutputParser
from langchain_openai import ChatOpenAI
from openai import OpenAI
import json

from config.settings import config
from prompts.load_prompts import load_prompt_from_json
from utils.serialize import serialize_chunk_langchain


class ChatLLM:
  def __init__(self, model, temperature=0.7):
    # 初始化模型设置
    self.llm = ChatOpenAI(
      model = model,
      temperature = temperature,
      api_key = config.LLM_API_KEY,
      base_url = config.LLM_BASE_URL,
      stream_usage=True
    )

    # 提示词模板
    prompt = load_prompt_from_json("src/prompts/prompts.json", 'chat_prompts')

    # 构建链
    self.chain = prompt | self.llm

  def generate_response(self, birth_date, birth_time):
    for chunk in self.chain.stream({"BIRTH_DATE": birth_date, "BIRTH_TIME": birth_time}):
      print(chunk.content, end="", flush=True)

      if chunk.usage_metadata:
        input_tokens = chunk.usage_metadata.get("input_tokens", 0)
        output_tokens = chunk.usage_metadata.get("output_tokens", 0)
        total_tokens = chunk.usage_metadata.get("total_tokens", 0)
    
    print(
      "\n"
      f"输入消耗token：{input_tokens}\n"
      f"输出消耗token：{output_tokens}\n"
      f"总计消耗token：{total_tokens}\n"
    )

  def test(self, question):
    prompt = load_prompt_from_json("src/prompts/prompts.json", 'test_prompts')
    chain = prompt | self.llm
    with open('src/langchain.json', "w", encoding="utf-8") as f:
      for chunk in chain.stream({"question": question}):
        chunk_dict = serialize_chunk_langchain(chunk)
        json.dump(chunk_dict, f, ensure_ascii=False)
        print(chunk.content, end="", flush=True)

        if chunk.usage_metadata:
          input_tokens = chunk.usage_metadata.get("input_tokens", 0)
          output_tokens = chunk.usage_metadata.get("output_tokens", 0)
          total_tokens = chunk.usage_metadata.get("total_tokens", 0)
    
    print(
      "\n"
      f"输入消耗token：{input_tokens}\n"
      f"输出消耗token：{output_tokens}\n"
      f"总计消耗token：{total_tokens}\n"
    )


if __name__=='__main__':
  llm = ChatLLM(model=config.MODEL)
  llm.generate_response("2001-10-20", "13:25")
  # llm.test("你是谁？")