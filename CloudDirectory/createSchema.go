package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/clouddirectory"
)

func main() {

	// Static credentials from environment variables
	id := os.Getenv("AWS_ACCESS_KEY_ID")
	secret := os.Getenv("AWS_SECRET_ACCESS_KEY")
	credentials := credentials.NewStaticCredentials(id, secret, "")

	// AWS session
	region := os.Getenv("AWS_REGION")
	session := session.Must(session.NewSession(
		&aws.Config{
			Region:      aws.String(region),
			Credentials: credentials,
			MaxRetries:  aws.Int(5),
		},
	))

	// Cloud Directory session
	cloudDirectorySession := clouddirectory.New(session)

	// Create the schema
	createSchemaInput := &clouddirectory.CreateSchemaInput{
		Name: aws.String("MyNewSchema"),
	}

	createSchamaOutput, err := cloudDirectorySession.CreateSchema(createSchemaInput)
	if err != nil {
		panic(err)
	}

	// read the schema
	data, err := ioutil.ReadFile("MySchema.json")
	if err != nil {
		panic(err)
	}
	schemaData := string(data)

	// upload the schema data
	putSchemaInput := &clouddirectory.PutSchemaFromJsonInput{
		SchemaArn: createSchamaOutput.SchemaArn,
		Document:  &schemaData,
	}
	putSchemaOutput, err := cloudDirectorySession.PutSchemaFromJson(putSchemaInput)
	if err != nil {
		panic(err)
	}

	fmt.Println(putSchemaOutput)

	// publish the schema
	publishSchemaInput := &clouddirectory.PublishSchemaInput{
		Name:                 aws.String("MyNewSchema"),
		DevelopmentSchemaArn: putSchemaOutput.Arn,
		Version:              aws.String("1.0"),
	}

	publishSchemaOutput, err := cloudDirectorySession.PublishSchema(publishSchemaInput)
	if err != nil {
		panic(err)
	}

	fmt.Println(publishSchemaOutput)
}
