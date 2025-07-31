package pbencoding

import (
	"errors"
	"fmt"
	"sentinel/packages/common/encoding"
	pbgen "sentinel/packages/common/proto/generated"

	"google.golang.org/protobuf/proto"
)

func marshall(pb any) ([]byte, error){
	encoding.Log.Trace("Marshalling data into a protobuf...", nil)

	var message proto.Message

	switch v := pb.(type) {
	case
		// user
		*pbgen.PublicUserDTO, *pbgen.BasicUserDTO, *pbgen.FullUserDTO, *pbgen.AuditUserDTO,
		// session
		*pbgen.PublicSessionDTO, *pbgen.FullSessionDTO,
		// location
		*pbgen.FullLocationDTO:
		message = v.(proto.Message)
	default:
		errMsg := fmt.Sprintf("Unexpected type: %T", v)
		encoding.Log.Error("Failed to marshall data into protobuf", errMsg, nil)
		return nil, errors.New(errMsg)
	}

	data, err := proto.Marshal(message)
	if err != nil {
		encoding.Log.Error("Failed to marshall data into protobuf", err.Error(), nil)
		return nil, errors.New("Faield to marshall data into protobuf: " + err.Error())
	}

	encoding.Log.Trace("Marshalling data into a protobuf: OK", nil)

	return data, nil
}

func unmarshall[T proto.Message](message T, rawDTO []byte) (T, error) {
	encoding.Log.Trace("Unmarshalling data from protobuf...", nil)

	var zero T

	if err := proto.Unmarshal(rawDTO, message); err != nil {
		encoding.Log.Error("Failed to unmarshall data from protobuf", err.Error(), nil)
		return zero, errors.New("Faield to unmarshall data from protobuf: " + err.Error())
	}

	encoding.Log.Trace("Unmarshalling data from protobuf: OK", nil)

	return message, nil
}

