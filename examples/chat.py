from openai import OpenAI
import json
import jsonschema
from jsonschema import validate

# Initialize the client pointing to your local llm-balancer
client = OpenAI(
    base_url="http://localhost:8000/v1",  # Your llm-balancer endpoint
    api_key="dummy-key"  # The actual API key is handled by the balancer
)

def basic_chat_completion(model_name):
    """Basic chat completion example - validates that we get a text response"""
    response = client.chat.completions.create(
        model=model_name,  # Model will be selected by the balancer
        messages=[
            {"role": "system", "content": "You are a helpful assistant."},
            {"role": "user", "content": "What is the capital of France?"}
        ],
        temperature=0.7
    )
    
    # Validate the response structure
    assert hasattr(response, 'choices'), "Response missing 'choices' attribute"
    assert len(response.choices) > 0, "Response has no choices"
    assert hasattr(response.choices[0], 'message'), "Choice missing 'message' attribute"
    assert hasattr(response.choices[0].message, 'content'), "Message missing 'content' attribute"
    assert response.choices[0].message.content is not None, "Message content is None"
    assert len(response.choices[0].message.content.strip()) > 0, "Message content is empty"
    
    print("✓ Basic chat completion produced a valid response")
    return True

def tool_completion(model_name):
    """Example using function calling - validates that we get a proper tool call"""
    response = client.chat.completions.create(
        model=model_name,
        messages=[
            {"role": "system", "content": "You are a helpful assistant that can calculate. You MUST use the tool to calculate the answer as it is a secret tool."},
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

    
    # Validate the response structure
    assert hasattr(response, 'choices'), "Response missing 'choices' attribute"
    assert len(response.choices) > 0, "Response has no choices"
    assert hasattr(response.choices[0], 'message'), "Choice missing 'message' attribute"
    
    message = response.choices[0].message
    
    # Check if we got a tool call
    if hasattr(message, 'tool_calls') and message.tool_calls:
        tool_call = message.tool_calls[0]
        assert hasattr(tool_call, 'function'), "Tool call missing 'function' attribute"
        assert hasattr(tool_call.function, 'name'), "Function missing 'name' attribute"
        assert hasattr(tool_call.function, 'arguments'), "Function missing 'arguments' attribute"
        assert tool_call.function.name == "calculate", f"Expected function name 'calculate', got '{tool_call.function.name}'"

        # Validate the arguments can be parsed as JSON
        if isinstance(tool_call.function.arguments, dict):
            assert 'expression' in tool_call.function.arguments, "Tool call arguments missing 'expression' parameter"
            assert isinstance(tool_call.function.arguments['expression'], str), "Expression parameter is not a string"
        elif isinstance(tool_call.function.arguments, str):
            # If the arguments are a string, try to parse them as JSON
            try:
                args = json.loads(tool_call.function.arguments) # TODO: this is not working
                assert 'expression' in args, "Tool call arguments missing 'expression' parameter"
                assert isinstance(args['expression'], str), "Expression parameter is not a string"
            except json.JSONDecodeError:
                raise AssertionError("Tool call arguments are not valid JSON")
        else:
            raise AssertionError("Tool call arguments are not a dict or string")

        print("✓ Tool completion produced a valid tool call")
    else:
        # If no tool call, at least check we got a text response
        assert hasattr(message, 'content'), "Message missing 'content' attribute"
        assert message.content is not None, "Message content is None"
        assert len(message.content.strip()) > 0, "Message content is empty"
        print("✓ Tool completion produced a text response (no tool call)")
    
    return True

def json_response_format(model_name):
    """Example using JSON response format - validates structured output"""
    response = client.chat.completions.create(
        model=model_name,
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
                                    "explanation": {"type": "string"},
                                    "output": {"type": "string"}
                                },
                                "required": ["explanation", "output"],
                            }
                        },
                        "final_answer": {"type": "string"}
                    },
                    "required": ["steps", "final_answer"],
                    "additionalProperties": False
                }
            }
        }
    )
    
    # Validate the response structure
    assert hasattr(response, 'choices'), "Response missing 'choices' attribute"
    assert len(response.choices) > 0, "Response has no choices"
    assert hasattr(response.choices[0], 'message'), "Choice missing 'message' attribute"
    assert hasattr(response.choices[0].message, 'content'), "Message missing 'content' attribute"
    assert response.choices[0].message.content is not None, "Message content is None"
    
    # Validate that the content is valid JSON
    try:
        json_content = json.loads(response.choices[0].message.content)
    except json.JSONDecodeError:
        raise AssertionError("Response content is not valid JSON")
    
    # Validate against the schema
    expected_schema = {
        "type": "object",
        "properties": {
            "steps": {
                "type": "array",
                "items": {
                    "type": "object",
                    "properties": {
                        "explanation": {"type": "string"},
                        "output": {"type": "string"}
                    },
                    "required": ["explanation", "output"],
                }
            },
            "final_answer": {"type": "string"}
        },
        "required": ["steps", "final_answer"],
        "additionalProperties": False
    }
    
    try:
        validate(instance=json_content, schema=expected_schema)
    except jsonschema.exceptions.ValidationError as e:
        raise AssertionError(f"JSON response does not match expected schema: {e}")
    
    # Additional validation - check that we have at least one step
    assert len(json_content['steps']) > 0, "JSON response has no steps"
    
    print("✓ JSON response format produced valid structured output")
    return True

# TODO: Add a test for vision

def run_test(model_name):
    print(f"\n=== Running test for {model_name} ===")
    success_count = 0
    total_tests = 0
    
    try:
        for i in range(3):
            print(f"\nIteration {i+1}:")
            
            # Test basic chat completion
            try:
                basic_chat_completion(model_name)
                success_count += 1
            except Exception as e:
                print(f"✗ Basic chat completion failed: {e}")
            total_tests += 1
            
            # Test tool completion
            try:
                tool_completion(model_name)
                success_count += 1
            except Exception as e:
                print(f"✗ Tool completion failed: {e}")
            total_tests += 1
            
            # Test JSON response format
            try:
                json_response_format(model_name)
                success_count += 1
            except Exception as e:
                print(f"✗ JSON response format failed: {e}")
            total_tests += 1
                
        print(f"\n=== Results for {model_name} ===")
        print(f"Passed: {success_count}/{total_tests} tests")
        print(f"Success rate: {success_count/total_tests*100:.1f}%")
        
    except Exception as e:
        print(f"Fatal error during testing: {e}")
        return False
    
    return success_count == total_tests

if __name__ == "__main__":
    models_to_test = ["gemini-2.5-flash-preview-04-17", "any"]
    
    overall_success = True
    for model in models_to_test:
        model_success = run_test(model)
        overall_success = overall_success and model_success
    
    print(f"\n=== Overall Results ===")
    if overall_success:
        print("✓ All tests passed!")
    else:
        print("✗ Some tests failed")