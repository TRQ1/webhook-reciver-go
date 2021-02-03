package lambda

import (
        "fmt"
        "context"
        "github.com/aws/aws-lambda-go/lambda"
)

type webHook struct {
        Name string `json:"name"`
}

func HandleRequest(ctx context.Context, name webHook) (string, error) {
        return fmt.Sprintf("Hello %s!", name.Name ), nil
}

func main() {
        lambda.Start(HandleRequest)
}