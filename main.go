package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"unicode"
)

const mainPageTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>🌸 Art Decoder/Encoder 🌼</title>
    <style>
        body {
						font-family: 'Quicksand', sans-serif;
            background-color: #87CEEB; /* Softer background */
            margin: 0;
            padding: 20px;
            color: #333; /* Dimmed text color */
        }
        .container {
            max-width: 600px;
            margin: auto;
            background: #ffffff; /* Lighter container background for contrast */
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2); /* Softer shadow */
        }
        h1, h2 {
            text-align: center;
            color: #575757;
        }
        form {
            margin-top: 20px;
        }
        textarea, select {
            width: 100%;
            padding: 10px;
            margin-bottom: 20px;
            border-radius: 4px;
            border: 1px solid #d3d3d3; /* Softer border */
            box-sizing: border-box;
            background-color: #ffffff; /* Ensure readability */
            color: #575757; /* Text color for inputs */
        }
        .button {
            display: block;
            width: 100%;
            padding: 10px;
            background-color: #FFEB3B; /* Muted button color */
            color: white;
            border: none;
            border-radius: 4px;
            cursor: pointer;
						transition: transform 0.2s;
        }
        .button:hover {
            background-color: #5a6268; /* Darken button on hover for feedback */
						transform: scale(1.05);
        }
        .result {
            margin-top: 20px;
            background-color: #e1f5fe; /* Soft background for result */
            padding: 10px;
            border-radius: 4px;
            color: #575757; /* Dimmed text for result */
						
        }
        .error {
            color: #a94442; /* Softer error color */
            text-align: center;
            margin-top: 20px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🌸 Art Decoder/Encoder 🌼</h1>
        <form action="/decoder" method="post">
            <textarea name="data" placeholder="Enter your text here..."></textarea>
            <select name="action">
                <option value="decode">Decode</option>
                <option value="encode">Encode</option>
            </select>
            <button class="button" type="submit">Submit</button>
        </form>
        {{if .Error}}
            <div class="error">{{.Error}}</div>
        {{end}}
        {{if .Result}}
            <div class="result">
                <h2>Result:</h2>
                <pre>{{.Result}}</pre>
            </div>
        {{end}}
    </div>
</body>
</html>
`

var tmpl = template.Must(template.New("mainPage").Parse(mainPageTemplate))

func mainPageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl.Execute(w, nil)
		fmt.Println("GET / - HTTP 200 OK") // Log the HTTP 200 status
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		fmt.Println("HTTP 405 Method Not Allowed") // Log when a non-GET request is made
	}
}

func decoderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		tmpl.Execute(w, map[string]interface{}{"Error": "Method not allowed"})
		return
	}

	data := r.FormValue("data")
	action := r.FormValue("action")
	var result string
	var hasError bool

	switch action {
	case "encode":
		result = encodeLine(data)
		fmt.Println("POST /decoder (encode) - HTTP 202 Accepted")
	case "decode":
		result, hasError = decodeLine(data)
		if hasError {
			fmt.Println("POST /decoder (decode) - Malformed encoded string")
			tmpl.Execute(w, map[string]interface{}{"Error": "Malformed encoded string"})
			return
		}
		fmt.Println("POST /decoder (decode) - HTTP 202 Accepted")
	default:
		tmpl.Execute(w, map[string]interface{}{"Error": "Invalid action"})
		return
	}

	tmpl.Execute(w, map[string]interface{}{"Result": result})
}

func decodeLine(inputstr string) (string, bool) {
	lines := strings.Split(inputstr, "\n")
	var output strings.Builder
	for _, line := range lines {
		decoded, hasError := processLine(line)
		if hasError {
			return "", true
		}
		output.WriteString(decoded + "\n")
	}
	return output.String(), false
}

func encodeLine(inputstr string) string {
	var prevChar string
	var count int
	var result strings.Builder

	for i := 0; i < len(inputstr); i++ {
		char := string(inputstr[i])
		if char == prevChar {
			count++
		} else {
			if count > 1 {
				result.WriteString(fmt.Sprintf("[%d %s]", count, prevChar))
			} else if count == 1 {
				result.WriteString(prevChar)
			}
			prevChar = char
			count = 1
		}
	}

	if count > 1 {
		result.WriteString(fmt.Sprintf("[%d %s]", count, prevChar))
	} else if count == 1 {
		result.WriteString(prevChar)
	}

	return result.String()
}

func processLine(line string) (string, bool) {
	var output strings.Builder
	i := 0

	for i < len(line) {
		if line[i] == '[' {
			countStart := i + 1
			bracketCount := 1
			i++

			for i < len(line) && bracketCount > 0 {
				if line[i] == '[' {
					bracketCount++
				} else if line[i] == ']' {
					bracketCount--
				}
				i++
			}

			if bracketCount != 0 { // Mismatched brackets
				return "", true
			}

			countAndChars := line[countStart : i-1]
			parts := splitCountAndChars(countAndChars)
			if len(parts) != 2 || parts[1] == "" || !startsWithNumber(parts[0]) {
				return "", true
			}

			count, err := strconv.Atoi(parts[0])
			if err != nil {
				return "", true
			}

			chars, hasError := processLine(parts[1]) // Recursively process nested structures
			if hasError {
				return "", true
			}

			for j := 0; j < count; j++ {
				output.WriteString(chars)
			}

		} else {
			output.WriteString(string(line[i]))
			i++
		}
	}

	return output.String(), false
}

func splitCountAndChars(s string) []string {
	var parts []string
	for i := 0; i < len(s); i++ {
		if s[i] == ' ' {
			parts = append(parts, s[:i], s[i+1:])
			break
		}
	}
	if len(parts) == 0 {
		parts = append(parts, s)
	}
	return parts
}

func startsWithNumber(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func main() {
	http.HandleFunc("/", mainPageHandler)
	http.HandleFunc("/decoder", decoderHandler)
	fmt.Println("Server starting on http://localhost:8080/")
	http.ListenAndServe(":8080", nil)
}
