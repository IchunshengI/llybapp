import json
from langchain_core.prompts import ChatPromptTemplate

def load_prompt_from_json(file_path: str, select_prompt: str = 'test_prompts'):
  """
  读取json内的模板内容。
  - file_path: 'json文件的路径'
  - select_prompt: '所选择的prompt模板'
    1. test_prompts (default)
    2. chat_prompts
  """
  # 打开JSON文件
  with open(file_path, 'r', encoding='utf-8') as f:
    prompt = json.load(f)

  # 获取chat_prompt的模板
  chat_prompt = prompt.get(select_prompt, [])

  # 创建PromptTemplate
  return ChatPromptTemplate.from_messages(chat_prompt)

# prompt_template = load_prompt_from_json("src/prompts/prompts.json", select_prompt='chat_prompts')
# print(prompt_template.format(BIRTH_DATE='2001-10-20', BIRTH_TIME='13:25'))