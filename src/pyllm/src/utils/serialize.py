def serialize_chunk_langchain(chunk):
  return chunk.__dict__ if hasattr(chunk, '__dict__') else str(chunk)