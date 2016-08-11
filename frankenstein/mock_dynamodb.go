// Copyright 2016 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package frankenstein

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type mockDynamoDB struct {
	tables map[string]mockDynamoDBTable
}

type mockDynamoDBTable struct {
	hashKey  string
	rangeKey string
	items    map[string][]mockDynamoDBItem
}

type mockDynamoDBItem map[string]*dynamodb.AttributeValue

func newMockDynamoDB() *mockDynamoDB {
	return &mockDynamoDB{
		tables: map[string]mockDynamoDBTable{},
	}
}

func (m *mockDynamoDB) CreateTable(input *dynamodb.CreateTableInput) (*dynamodb.CreateTableOutput, error) {
	var hashKey, rangeKey string
	for _, schemaElement := range input.KeySchema {
		if *schemaElement.KeyType == "HASH" {
			hashKey = *schemaElement.AttributeName
		} else if *schemaElement.KeyType == "RANGE" {
			rangeKey = *schemaElement.AttributeName
		}
	}

	m.tables[*input.TableName] = mockDynamoDBTable{
		hashKey:  hashKey,
		rangeKey: rangeKey,
		items:    map[string][]mockDynamoDBItem{},
	}

	return &dynamodb.CreateTableOutput{}, nil
}

func (m *mockDynamoDB) ListTables(*dynamodb.ListTablesInput) (*dynamodb.ListTablesOutput, error) {
	var tableNames []*string
	for tableName, _ := range m.tables {
		func(tableName string) {
			tableNames = append(tableNames, &tableName)
		}(tableName)
	}
	return &dynamodb.ListTablesOutput{
		TableNames: tableNames,
	}, nil
}

func (m *mockDynamoDB) BatchWriteItem(input *dynamodb.BatchWriteItemInput) (*dynamodb.BatchWriteItemOutput, error) {
	for tableName, writeRequests := range input.RequestItems {
		table, ok := m.tables[tableName]
		if !ok {
			return &dynamodb.BatchWriteItemOutput{}, fmt.Errorf("table not found")
		}

		for _, writeRequest := range writeRequests {
			hashValue := *writeRequest.PutRequest.Item[table.hashKey].S
			rangeValue := writeRequest.PutRequest.Item[table.rangeKey].B
			fmt.Printf("Write %s/%x\n", hashValue, rangeValue)

			items := table.items[hashValue]

			// insert in order
			i := sort.Search(len(items), func(i int) bool {
				return bytes.Compare(items[i][table.rangeKey].B, rangeValue) >= 0
			})
			if i >= len(items) || !bytes.Equal(items[i][table.rangeKey].B, rangeValue) {
				items = append(items, nil)
				copy(items[i+1:], items[i:])
			}
			items[i] = writeRequest.PutRequest.Item

			table.items[hashValue] = items
		}
	}
	return &dynamodb.BatchWriteItemOutput{}, nil
}

func (m *mockDynamoDB) Query(input *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
	table, ok := m.tables[*input.TableName]
	if !ok {
		return nil, fmt.Errorf("table not found")
	}

	hashValue := *input.KeyConditions[table.hashKey].AttributeValueList[0].S
	items, ok := table.items[hashValue]
	if !ok {
		return &dynamodb.QueryOutput{}, nil
	}

	rangeValueStart := input.KeyConditions[table.rangeKey].AttributeValueList[0].B
	rangeValueEnd := input.KeyConditions[table.rangeKey].AttributeValueList[1].B

	fmt.Printf("Lookup %s/%x -> %x\n", hashValue, rangeValueStart, rangeValueEnd)

	i := sort.Search(len(items), func(i int) bool {
		return bytes.Compare(items[i][table.rangeKey].B, rangeValueStart) >= 0
	})
	if i >= len(items) {
		return &dynamodb.QueryOutput{}, nil
	}

	j := sort.Search(len(items), func(i int) bool {
		return bytes.Compare(items[i][table.rangeKey].B, rangeValueEnd) >= 0
	})

	result := make([]map[string]*dynamodb.AttributeValue, 0, j-i)
	for _, item := range items[i : j-i] {
		result = append(result, item)
	}

	return &dynamodb.QueryOutput{
		Items: result,
	}, nil
}
