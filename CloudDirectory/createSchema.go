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

	createSchemaOutput, err := cloudDirectorySession.CreateSchema(createSchemaInput)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Created new schema: %v\n", createSchemaOutput)

	// read the schema
	data, err := ioutil.ReadFile("MySchema.json")
	if err != nil {
		panic(err)
	}
	schemaData := string(data)

	// upload the schema data
	putSchemaInput := &clouddirectory.PutSchemaFromJsonInput{
		SchemaArn: createSchemaOutput.SchemaArn,
		Document:  &schemaData,
	}
	putSchemaOutput, err := cloudDirectorySession.PutSchemaFromJson(putSchemaInput)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Added schema data: %v\n", putSchemaOutput)

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

	fmt.Printf("Published schema: %v\n", publishSchemaOutput)

	// Create a directory
	createDirectoryInput := &clouddirectory.CreateDirectoryInput{
		Name:      aws.String("MyDirectory"),
		SchemaArn: publishSchemaOutput.PublishedSchemaArn,
	}

	createDirectoryOutput, err := cloudDirectorySession.CreateDirectory(createDirectoryInput)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Created new directory: %v\n", createDirectoryOutput)

	// Create an Index holder underthe root
	createBranchInput := clouddirectory.CreateObjectInput{
		DirectoryArn: createDirectoryOutput.DirectoryArn,
		ParentReference: &clouddirectory.ObjectReference{
			Selector: aws.String("/"),
		},
		LinkName: aws.String("indices"),
		SchemaFacets: []*clouddirectory.SchemaFacet{
			{
				FacetName: aws.String("indices"),
				SchemaArn: createDirectoryOutput.AppliedSchemaArn,
			},
		},
	}
	createBranchOutput, err := cloudDirectorySession.CreateObject(&createBranchInput)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Created new branch: %v\n", createBranchOutput)

	// Create an index and hang under the holder
	createIndexInput := clouddirectory.CreateIndexInput{
		DirectoryArn: createDirectoryOutput.DirectoryArn,
		IsUnique:     aws.Bool(true),
		LinkName:     aws.String("org_index"),
		OrderedIndexedAttributeList: []*clouddirectory.AttributeKey{
			{
				FacetName: aws.String("organization"),
				Name:      aws.String("name"),
				SchemaArn: createDirectoryOutput.AppliedSchemaArn,
			},
		},
		ParentReference: &clouddirectory.ObjectReference{
			Selector: aws.String("/indices"),
		},
	}

	createIndexOutput, err := cloudDirectorySession.CreateIndex(&createIndexInput)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Created new Index: %v\n", createIndexOutput)

	// Create an object
	createOrganizationInput := clouddirectory.CreateObjectInput{
		DirectoryArn: createDirectoryOutput.DirectoryArn,
		ParentReference: &clouddirectory.ObjectReference{
			Selector: aws.String("/"),
		},
		LinkName: aws.String("MyOrganization"),
		SchemaFacets: []*clouddirectory.SchemaFacet{
			{
				FacetName: aws.String("organization"),
				SchemaArn: createDirectoryOutput.AppliedSchemaArn,
			},
		},
		ObjectAttributeList: []*clouddirectory.AttributeKeyAndValue{
			&clouddirectory.AttributeKeyAndValue{
				Key: &clouddirectory.AttributeKey{
					FacetName: aws.String("organization"),
					Name:      aws.String("name"),
					SchemaArn: createDirectoryOutput.AppliedSchemaArn,
				},
				Value: &clouddirectory.TypedAttributeValue{
					StringValue: aws.String("MySampleOrg"),
				},
			},
		},
	}
	createOrganizationOutput, err := cloudDirectorySession.CreateObject(&createOrganizationInput)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Created new object: %v\n", createOrganizationOutput)

	// Attach the object to the index
	attachToIndexInput := &clouddirectory.AttachToIndexInput{
		DirectoryArn: createDirectoryOutput.DirectoryArn,
		IndexReference: &clouddirectory.ObjectReference{
			Selector: aws.String("$" + *createIndexOutput.ObjectIdentifier),
		},
		TargetReference: &clouddirectory.ObjectReference{
			Selector: aws.String("$" + *createOrganizationOutput.ObjectIdentifier),
		},
	}

	attachToIndexOutput, err := cloudDirectorySession.AttachToIndex(attachToIndexInput)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Attached object to index: %v\n", attachToIndexOutput)
}
