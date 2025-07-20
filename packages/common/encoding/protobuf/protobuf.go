package pbencoding

import (
	"errors"
	"fmt"
	"sentinel/packages/common/encoding"
	pbgen "sentinel/packages/common/proto/generated"

	"google.golang.org/protobuf/proto"
)

func marshall(pb any) ([]byte, error){
	encoding.Logger.Trace("Marshalling user DTO into a protobuf...", nil)

	var message proto.Message

	switch v := pb.(type) {
	case
		// user
		*pbgen.PublicUserDTO, *pbgen.BasicUserDTO, *pbgen.ExtendedUserDTO, *pbgen.AuditUserDTO,
		// session
		*pbgen.PublicSessionDTO, *pbgen.FullSessionDTO,
		// location
		*pbgen.FullLocationDTO:
		message = v.(proto.Message)
	default:
		return nil, fmt.Errorf("TYPE ERROR: Failed to marshall data into protobuf. Unexpected type: %T", v)
	}

	data, err := proto.Marshal(message)
	if err != nil {
		return nil, errors.New("Faield to marshall data in protobuf: " + err.Error())
	}

	encoding.Logger.Trace("Marshalling data into a protobuf: OK", nil)

	return data, nil
}

func unmarshall[T proto.Message](message T, rawDTO []byte) (T, error) {
	encoding.Logger.Trace("Unmarshalling data from protobuf...", nil)

	var zero T

	if err := proto.Unmarshal(rawDTO, message); err != nil {
		return zero, errors.New("Faield to unmarshall data from protobuf: " + err.Error())
	}

	encoding.Logger.Trace("Unmarshalling data from protobuf: OK", nil)

	return message, nil
}

