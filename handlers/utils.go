package handlers

import "github.com/pkoukk/tiktoken-go"

func countTokens(text string) (int, error) {
	encoding, err := tiktoken.GetEncoding("cl100k_base") // GPT-4/GPT-3.5-turbo encoding
	if err != nil {
		return 0, err
	}

	tokens := encoding.Encode(text, nil, nil)
	return len(tokens), nil
}
