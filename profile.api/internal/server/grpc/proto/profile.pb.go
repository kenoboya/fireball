// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v6.31.0--rc2
// source: internal/server/grpc/proto/profile.proto

package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Response struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Success       bool                   `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	Message       string                 `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Response) Reset() {
	*x = Response{}
	mi := &file_internal_server_grpc_proto_profile_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Response) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Response) ProtoMessage() {}

func (x *Response) ProtoReflect() protoreflect.Message {
	mi := &file_internal_server_grpc_proto_profile_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Response.ProtoReflect.Descriptor instead.
func (*Response) Descriptor() ([]byte, []int) {
	return file_internal_server_grpc_proto_profile_proto_rawDescGZIP(), []int{0}
}

func (x *Response) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

func (x *Response) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

type UserRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	SenderID      string                 `protobuf:"bytes,1,opt,name=senderID,proto3" json:"senderID,omitempty"`
	RecipientID   string                 `protobuf:"bytes,2,opt,name=recipientID,proto3" json:"recipientID,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UserRequest) Reset() {
	*x = UserRequest{}
	mi := &file_internal_server_grpc_proto_profile_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UserRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UserRequest) ProtoMessage() {}

func (x *UserRequest) ProtoReflect() protoreflect.Message {
	mi := &file_internal_server_grpc_proto_profile_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UserRequest.ProtoReflect.Descriptor instead.
func (*UserRequest) Descriptor() ([]byte, []int) {
	return file_internal_server_grpc_proto_profile_proto_rawDescGZIP(), []int{1}
}

func (x *UserRequest) GetSenderID() string {
	if x != nil {
		return x.SenderID
	}
	return ""
}

func (x *UserRequest) GetRecipientID() string {
	if x != nil {
		return x.RecipientID
	}
	return ""
}

type UsersRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	SenderID      string                 `protobuf:"bytes,1,opt,name=senderID,proto3" json:"senderID,omitempty"`
	RecipientIDs  []string               `protobuf:"bytes,2,rep,name=recipientIDs,proto3" json:"recipientIDs,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UsersRequest) Reset() {
	*x = UsersRequest{}
	mi := &file_internal_server_grpc_proto_profile_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UsersRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UsersRequest) ProtoMessage() {}

func (x *UsersRequest) ProtoReflect() protoreflect.Message {
	mi := &file_internal_server_grpc_proto_profile_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UsersRequest.ProtoReflect.Descriptor instead.
func (*UsersRequest) Descriptor() ([]byte, []int) {
	return file_internal_server_grpc_proto_profile_proto_rawDescGZIP(), []int{2}
}

func (x *UsersRequest) GetSenderID() string {
	if x != nil {
		return x.SenderID
	}
	return ""
}

func (x *UsersRequest) GetRecipientIDs() []string {
	if x != nil {
		return x.RecipientIDs
	}
	return nil
}

type User struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	UserID        string                 `protobuf:"bytes,1,opt,name=userID,proto3" json:"userID,omitempty"`
	Username      string                 `protobuf:"bytes,2,opt,name=username,proto3" json:"username,omitempty"`
	DisplayName   string                 `protobuf:"bytes,3,opt,name=displayName,proto3" json:"displayName,omitempty"`
	Bio           *string                `protobuf:"bytes,4,opt,name=bio,proto3,oneof" json:"bio,omitempty"`
	Email         *string                `protobuf:"bytes,5,opt,name=email,proto3,oneof" json:"email,omitempty"`
	Phone         *string                `protobuf:"bytes,6,opt,name=phone,proto3,oneof" json:"phone,omitempty"`
	AvatarURL     *string                `protobuf:"bytes,7,opt,name=avatarURL,proto3,oneof" json:"avatarURL,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *User) Reset() {
	*x = User{}
	mi := &file_internal_server_grpc_proto_profile_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *User) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*User) ProtoMessage() {}

func (x *User) ProtoReflect() protoreflect.Message {
	mi := &file_internal_server_grpc_proto_profile_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use User.ProtoReflect.Descriptor instead.
func (*User) Descriptor() ([]byte, []int) {
	return file_internal_server_grpc_proto_profile_proto_rawDescGZIP(), []int{3}
}

func (x *User) GetUserID() string {
	if x != nil {
		return x.UserID
	}
	return ""
}

func (x *User) GetUsername() string {
	if x != nil {
		return x.Username
	}
	return ""
}

func (x *User) GetDisplayName() string {
	if x != nil {
		return x.DisplayName
	}
	return ""
}

func (x *User) GetBio() string {
	if x != nil && x.Bio != nil {
		return *x.Bio
	}
	return ""
}

func (x *User) GetEmail() string {
	if x != nil && x.Email != nil {
		return *x.Email
	}
	return ""
}

func (x *User) GetPhone() string {
	if x != nil && x.Phone != nil {
		return *x.Phone
	}
	return ""
}

func (x *User) GetAvatarURL() string {
	if x != nil && x.AvatarURL != nil {
		return *x.AvatarURL
	}
	return ""
}

type UsersProfileResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Users         []*User                `protobuf:"bytes,1,rep,name=users,proto3" json:"users,omitempty"`
	Response      *Response              `protobuf:"bytes,2,opt,name=response,proto3" json:"response,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UsersProfileResponse) Reset() {
	*x = UsersProfileResponse{}
	mi := &file_internal_server_grpc_proto_profile_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UsersProfileResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UsersProfileResponse) ProtoMessage() {}

func (x *UsersProfileResponse) ProtoReflect() protoreflect.Message {
	mi := &file_internal_server_grpc_proto_profile_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UsersProfileResponse.ProtoReflect.Descriptor instead.
func (*UsersProfileResponse) Descriptor() ([]byte, []int) {
	return file_internal_server_grpc_proto_profile_proto_rawDescGZIP(), []int{4}
}

func (x *UsersProfileResponse) GetUsers() []*User {
	if x != nil {
		return x.Users
	}
	return nil
}

func (x *UsersProfileResponse) GetResponse() *Response {
	if x != nil {
		return x.Response
	}
	return nil
}

type UserBriefInfoResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	UserID        string                 `protobuf:"bytes,1,opt,name=userID,proto3" json:"userID,omitempty"`
	Username      string                 `protobuf:"bytes,2,opt,name=username,proto3" json:"username,omitempty"`
	Name          string                 `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	AvatarURL     *string                `protobuf:"bytes,4,opt,name=avatarURL,proto3,oneof" json:"avatarURL,omitempty"`
	Response      *Response              `protobuf:"bytes,5,opt,name=response,proto3" json:"response,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UserBriefInfoResponse) Reset() {
	*x = UserBriefInfoResponse{}
	mi := &file_internal_server_grpc_proto_profile_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UserBriefInfoResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UserBriefInfoResponse) ProtoMessage() {}

func (x *UserBriefInfoResponse) ProtoReflect() protoreflect.Message {
	mi := &file_internal_server_grpc_proto_profile_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UserBriefInfoResponse.ProtoReflect.Descriptor instead.
func (*UserBriefInfoResponse) Descriptor() ([]byte, []int) {
	return file_internal_server_grpc_proto_profile_proto_rawDescGZIP(), []int{5}
}

func (x *UserBriefInfoResponse) GetUserID() string {
	if x != nil {
		return x.UserID
	}
	return ""
}

func (x *UserBriefInfoResponse) GetUsername() string {
	if x != nil {
		return x.Username
	}
	return ""
}

func (x *UserBriefInfoResponse) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *UserBriefInfoResponse) GetAvatarURL() string {
	if x != nil && x.AvatarURL != nil {
		return *x.AvatarURL
	}
	return ""
}

func (x *UserBriefInfoResponse) GetResponse() *Response {
	if x != nil {
		return x.Response
	}
	return nil
}

type UsersBriefInfoResponse struct {
	state                  protoimpl.MessageState   `protogen:"open.v1"`
	UsersBriefInfoResponse []*UserBriefInfoResponse `protobuf:"bytes,1,rep,name=usersBriefInfoResponse,proto3" json:"usersBriefInfoResponse,omitempty"`
	Response               *Response                `protobuf:"bytes,2,opt,name=response,proto3" json:"response,omitempty"`
	unknownFields          protoimpl.UnknownFields
	sizeCache              protoimpl.SizeCache
}

func (x *UsersBriefInfoResponse) Reset() {
	*x = UsersBriefInfoResponse{}
	mi := &file_internal_server_grpc_proto_profile_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UsersBriefInfoResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UsersBriefInfoResponse) ProtoMessage() {}

func (x *UsersBriefInfoResponse) ProtoReflect() protoreflect.Message {
	mi := &file_internal_server_grpc_proto_profile_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UsersBriefInfoResponse.ProtoReflect.Descriptor instead.
func (*UsersBriefInfoResponse) Descriptor() ([]byte, []int) {
	return file_internal_server_grpc_proto_profile_proto_rawDescGZIP(), []int{6}
}

func (x *UsersBriefInfoResponse) GetUsersBriefInfoResponse() []*UserBriefInfoResponse {
	if x != nil {
		return x.UsersBriefInfoResponse
	}
	return nil
}

func (x *UsersBriefInfoResponse) GetResponse() *Response {
	if x != nil {
		return x.Response
	}
	return nil
}

var File_internal_server_grpc_proto_profile_proto protoreflect.FileDescriptor

const file_internal_server_grpc_proto_profile_proto_rawDesc = "" +
	"\n" +
	"(internal/server/grpc/proto/profile.proto\x12\x05proto\">\n" +
	"\bResponse\x12\x18\n" +
	"\asuccess\x18\x01 \x01(\bR\asuccess\x12\x18\n" +
	"\amessage\x18\x02 \x01(\tR\amessage\"K\n" +
	"\vUserRequest\x12\x1a\n" +
	"\bsenderID\x18\x01 \x01(\tR\bsenderID\x12 \n" +
	"\vrecipientID\x18\x02 \x01(\tR\vrecipientID\"N\n" +
	"\fUsersRequest\x12\x1a\n" +
	"\bsenderID\x18\x01 \x01(\tR\bsenderID\x12\"\n" +
	"\frecipientIDs\x18\x02 \x03(\tR\frecipientIDs\"\xf6\x01\n" +
	"\x04User\x12\x16\n" +
	"\x06userID\x18\x01 \x01(\tR\x06userID\x12\x1a\n" +
	"\busername\x18\x02 \x01(\tR\busername\x12 \n" +
	"\vdisplayName\x18\x03 \x01(\tR\vdisplayName\x12\x15\n" +
	"\x03bio\x18\x04 \x01(\tH\x00R\x03bio\x88\x01\x01\x12\x19\n" +
	"\x05email\x18\x05 \x01(\tH\x01R\x05email\x88\x01\x01\x12\x19\n" +
	"\x05phone\x18\x06 \x01(\tH\x02R\x05phone\x88\x01\x01\x12!\n" +
	"\tavatarURL\x18\a \x01(\tH\x03R\tavatarURL\x88\x01\x01B\x06\n" +
	"\x04_bioB\b\n" +
	"\x06_emailB\b\n" +
	"\x06_phoneB\f\n" +
	"\n" +
	"_avatarURL\"f\n" +
	"\x14UsersProfileResponse\x12!\n" +
	"\x05users\x18\x01 \x03(\v2\v.proto.UserR\x05users\x12+\n" +
	"\bresponse\x18\x02 \x01(\v2\x0f.proto.ResponseR\bresponse\"\xbd\x01\n" +
	"\x15UserBriefInfoResponse\x12\x16\n" +
	"\x06userID\x18\x01 \x01(\tR\x06userID\x12\x1a\n" +
	"\busername\x18\x02 \x01(\tR\busername\x12\x12\n" +
	"\x04name\x18\x03 \x01(\tR\x04name\x12!\n" +
	"\tavatarURL\x18\x04 \x01(\tH\x00R\tavatarURL\x88\x01\x01\x12+\n" +
	"\bresponse\x18\x05 \x01(\v2\x0f.proto.ResponseR\bresponseB\f\n" +
	"\n" +
	"_avatarURL\"\x9b\x01\n" +
	"\x16UsersBriefInfoResponse\x12T\n" +
	"\x16usersBriefInfoResponse\x18\x01 \x03(\v2\x1c.proto.UserBriefInfoResponseR\x16usersBriefInfoResponse\x12+\n" +
	"\bresponse\x18\x02 \x01(\v2\x0f.proto.ResponseR\bresponse2\xe4\x01\n" +
	"\x0eProfileService\x12D\n" +
	"\x10GetUserBriefInfo\x12\x12.proto.UserRequest\x1a\x1c.proto.UserBriefInfoResponse\x12G\n" +
	"\x11GetUsersBriefInfo\x12\x13.proto.UsersRequest\x1a\x1d.proto.UsersBriefInfoResponse\x12C\n" +
	"\x0fGetUsersProfile\x12\x13.proto.UsersRequest\x1a\x1b.proto.UsersProfileResponseB$Z\"./internal/server/grpc/proto;protob\x06proto3"

var (
	file_internal_server_grpc_proto_profile_proto_rawDescOnce sync.Once
	file_internal_server_grpc_proto_profile_proto_rawDescData []byte
)

func file_internal_server_grpc_proto_profile_proto_rawDescGZIP() []byte {
	file_internal_server_grpc_proto_profile_proto_rawDescOnce.Do(func() {
		file_internal_server_grpc_proto_profile_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_internal_server_grpc_proto_profile_proto_rawDesc), len(file_internal_server_grpc_proto_profile_proto_rawDesc)))
	})
	return file_internal_server_grpc_proto_profile_proto_rawDescData
}

var file_internal_server_grpc_proto_profile_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_internal_server_grpc_proto_profile_proto_goTypes = []any{
	(*Response)(nil),               // 0: proto.Response
	(*UserRequest)(nil),            // 1: proto.UserRequest
	(*UsersRequest)(nil),           // 2: proto.UsersRequest
	(*User)(nil),                   // 3: proto.User
	(*UsersProfileResponse)(nil),   // 4: proto.UsersProfileResponse
	(*UserBriefInfoResponse)(nil),  // 5: proto.UserBriefInfoResponse
	(*UsersBriefInfoResponse)(nil), // 6: proto.UsersBriefInfoResponse
}
var file_internal_server_grpc_proto_profile_proto_depIdxs = []int32{
	3, // 0: proto.UsersProfileResponse.users:type_name -> proto.User
	0, // 1: proto.UsersProfileResponse.response:type_name -> proto.Response
	0, // 2: proto.UserBriefInfoResponse.response:type_name -> proto.Response
	5, // 3: proto.UsersBriefInfoResponse.usersBriefInfoResponse:type_name -> proto.UserBriefInfoResponse
	0, // 4: proto.UsersBriefInfoResponse.response:type_name -> proto.Response
	1, // 5: proto.ProfileService.GetUserBriefInfo:input_type -> proto.UserRequest
	2, // 6: proto.ProfileService.GetUsersBriefInfo:input_type -> proto.UsersRequest
	2, // 7: proto.ProfileService.GetUsersProfile:input_type -> proto.UsersRequest
	5, // 8: proto.ProfileService.GetUserBriefInfo:output_type -> proto.UserBriefInfoResponse
	6, // 9: proto.ProfileService.GetUsersBriefInfo:output_type -> proto.UsersBriefInfoResponse
	4, // 10: proto.ProfileService.GetUsersProfile:output_type -> proto.UsersProfileResponse
	8, // [8:11] is the sub-list for method output_type
	5, // [5:8] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_internal_server_grpc_proto_profile_proto_init() }
func file_internal_server_grpc_proto_profile_proto_init() {
	if File_internal_server_grpc_proto_profile_proto != nil {
		return
	}
	file_internal_server_grpc_proto_profile_proto_msgTypes[3].OneofWrappers = []any{}
	file_internal_server_grpc_proto_profile_proto_msgTypes[5].OneofWrappers = []any{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_internal_server_grpc_proto_profile_proto_rawDesc), len(file_internal_server_grpc_proto_profile_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_internal_server_grpc_proto_profile_proto_goTypes,
		DependencyIndexes: file_internal_server_grpc_proto_profile_proto_depIdxs,
		MessageInfos:      file_internal_server_grpc_proto_profile_proto_msgTypes,
	}.Build()
	File_internal_server_grpc_proto_profile_proto = out.File
	file_internal_server_grpc_proto_profile_proto_goTypes = nil
	file_internal_server_grpc_proto_profile_proto_depIdxs = nil
}
