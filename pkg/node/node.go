/*
* Copyright 2020-present Arpabet Inc. All rights reserved.
 */


package node

import (
	c "context"
	"github.com/arpabet/sprint/pkg/app"
	"github.com/arpabet/sprint/pkg/pb"
	"github.com/arpabet/sprint/pkg/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io"
	"fmt"
)


func Dial() (*grpc.ClientConn, error) {

	config, err := util.LoadClientConfig()
	if err != nil {
		return nil, err
	}

	tlsCredentials := credentials.NewTLS(config)

	return grpc.Dial(app.GetNodeAddress(), grpc.WithTransportCredentials(tlsCredentials))
}

func Status() (string, error) {

	conn, err := Dial()
	if err != nil {
		return "", err
	}
	defer conn.Close()

	client := pb.NewNodeServiceClient(conn)

	if resp, err := client.Node(c.Background(), new(pb.NodeRequest)); err != nil {
		return "", err
	} else {
		return resp.String(), nil
	}
}

func Shutdown(restart bool) (string, error) {

	conn, err := Dial()
	if err != nil {
		return "", err
	}
	defer conn.Close()

	client := pb.NewNodeServiceClient(conn)

	req := &pb.ShutdownRequest {
		Restart: restart,
	}

	if _, err := client.Shutdown(c.Background(), req); err != nil {
		return "", err
	} else {
		return "SUCCESS", nil
	}
}

func SetConfig(key, value string) (string, error) {

	conn, err := Dial()
	if err != nil {
		return "", err
	}
	defer conn.Close()

	client := pb.NewNodeServiceClient(conn)

	request := &pb.SetConfigRequest{
		Key: key,
		Value: value,
	}

	if _, err := client.SetConfig(c.Background(), request); err != nil {
		return "", err
	} else {
		return "SUCCESS", nil
	}
}

func GetConfig(key string) (string, error) {

	conn, err := Dial()
	if err != nil {
		return "", err
	}
	defer conn.Close()

	client := pb.NewNodeServiceClient(conn)

	request := &pb.GetConfigRequest{
		Key: key,
	}

	if resp, err := client.GetConfig(c.Background(), request); err != nil {
		return "", err
	} else {
		if resp.Entry != nil {
			return resp.Entry.Value, nil
		} else {
			return "", nil
		}
	}

}

func GetConfiguration(writer io.StringWriter) error {

	conn, err := Dial()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pb.NewNodeServiceClient(conn)

	request := &pb.ConfigurationRequest{
	}

	if resp, err := client.Configuration(c.Background(), request); err != nil {
		return err
	} else {
		for _, entry := range resp.Entry {
			writer.WriteString(fmt.Sprintf("%s: %s\n", entry.Key, entry.Value))
		}
		return nil
	}

}


func DatabaseConsole(writer io.StringWriter, errWriter io.StringWriter) error {

	conn, err := Dial()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pb.NewNodeServiceClient(conn)

	stream, err := client.DatabaseConsole(c.Background())
	if err != nil {
		return err
	}

	barrier := make(chan int, 1)

	go func() {
		defer func() {
			barrier <- -1
		}()
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				errWriter.WriteString(fmt.Sprintf("error: recv i/o %v\n", err))
				break
			}
			switch resp.Status {
			case 100:
				barrier <- 1
			case 200:
				writer.WriteString(fmt.Sprintf("%s\n", resp.Content))
			default:
				errWriter.WriteString(fmt.Sprintf("error: code %d, %s\n", resp.Status, resp.Content))
			}
		}
	}()

	for {
		query := util.Prompt("Enter query [exit]: ")
		if query == "exit" {
			break
		}
		request := &pb.DatabaseRequest{
			Query:                query,
		}
		err = stream.Send(request)
		if err != nil {
			errWriter.WriteString(fmt.Sprintf("error: send i/o %v\n", err))
			break
		}
		if  <- barrier == -1 {
			break
		}
	}

	stream.CloseSend()
	return nil
}