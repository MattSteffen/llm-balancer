from openai import OpenAI
import json

# Initialize the client pointing to your local llm-balancer
client = OpenAI(
    base_url="http://localhost:8000/v1",  # Your llm-balancer endpoint
    api_key="dummy-key"  # The actual API key is handled by the balancer
)

def basic_chat_completion():
    """Basic chat completion example"""
    response = client.chat.completions.create(
        model="gemini-2.5-pro",  # Model will be selected by the balancer
        messages=[
            {"role": "system", "content": "You are a helpful assistant."},
            {"role": "user", "content": "What is the capital of France?"}
        ],
        temperature=0.7
    )
    print("\n=== Basic Chat Completion ===")
    print(response.choices[0].message.content)

def tool_completion():
    """Example using function calling"""
    response = client.chat.completions.create(
        model="gemini-2.5-pro",
        messages=[
            {"role": "system", "content": "You are a helpful assistant that can calculate."},
            {"role": "user", "content": "What is 23 plus 45?"}
        ],
        tools=[{
            "type": "function",
            "function": {
                "name": "calculate",
                "description": "Calculate a mathematical expression",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "expression": {
                            "type": "string",
                            "description": "The mathematical expression to calculate"
                        }
                    },
                    "required": ["expression"]
                }
            }
        }],
        tool_choice="auto"
    )
    print("\n=== Tool Completion ===")
    print(json.dumps(response.model_dump(), indent=2))

def json_response_format():
    """Example using JSON response format"""

    response = client.chat.completions.create(
        model="gemini-2.5-pro",
        messages=[
          {
            "role": "system",
            "content": "You are a helpful math tutor. Guide the user through the solution step by step."
          },
          {
            "role": "user",
            "content": "How can I solve 8x + 7 = -23?"
          }
        ],
        response_format={
          "type": "json_schema",
          "json_schema": {
            "name": "math_reasoning",
            "strict": True,
            "schema": {
              "type": "object",
              "properties": {
                "steps": {
                  "type": "array",
                  "items": {
                    "type": "object",
                    "properties": {
                      "explanation": { "type": "string" },
                      "output": { "type": "string" }
                    },
                    "required": ["explanation", "output"],
                    "additionalProperties": False
                  }
                },
                "final_answer": { "type": "string" }
              },
              "required": ["steps", "final_answer"],
              "additionalProperties": False
            }
          }
        }
    )
    print("\n=== JSON Response Format ===")
    print(json.dumps(json.loads(response.choices[0].message.content), indent=2))

if __name__ == "__main__":
    try:
        # basic_chat_completion()
        # tool_completion()
        for i in range(10):
            json_response_format()
    except Exception as e:
        print(f"Error: {e}")