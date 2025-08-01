// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.4.0
// - protoc             v5.29.3
// source: api/proto/job_service.proto

package job_service

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.62.0 or later.
const _ = grpc.SupportPackageIsVersion8

const (
	JobService_SubmitJob_FullMethodName       = "/job_service.JobService/SubmitJob"
	JobService_GetJobStatus_FullMethodName    = "/job_service.JobService/GetJobStatus"
	JobService_RegisterAgent_FullMethodName   = "/job_service.JobService/RegisterAgent"
	JobService_UpdateJobStatus_FullMethodName = "/job_service.JobService/UpdateJobStatus"
	JobService_FetchJob_FullMethodName        = "/job_service.JobService/FetchJob"
)

// JobServiceClient is the client API for JobService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// JobService is the main service for managing jobs.
type JobServiceClient interface {
	// SubmitJob submits a new test job.
	SubmitJob(ctx context.Context, in *SubmitJobRequest, opts ...grpc.CallOption) (*SubmitJobResponse, error)
	// GetJobStatus retrieves the status of a job.
	GetJobStatus(ctx context.Context, in *GetJobStatusRequest, opts ...grpc.CallOption) (*GetJobStatusResponse, error)
	// RegisterAgent allows an agent to register with the orchestrator.
	RegisterAgent(ctx context.Context, in *RegisterAgentRequest, opts ...grpc.CallOption) (*RegisterAgentResponse, error)
	// UpdateJobStatus is used by an agent to report job progress.
	UpdateJobStatus(ctx context.Context, in *UpdateJobStatusRequest, opts ...grpc.CallOption) (*UpdateJobStatusResponse, error)
	// FetchJob fetches a job for a given target capability.
	FetchJob(ctx context.Context, in *FetchJobRequest, opts ...grpc.CallOption) (*FetchJobResponse, error)
}

type jobServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewJobServiceClient(cc grpc.ClientConnInterface) JobServiceClient {
	return &jobServiceClient{cc}
}

func (c *jobServiceClient) SubmitJob(ctx context.Context, in *SubmitJobRequest, opts ...grpc.CallOption) (*SubmitJobResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SubmitJobResponse)
	err := c.cc.Invoke(ctx, JobService_SubmitJob_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *jobServiceClient) GetJobStatus(ctx context.Context, in *GetJobStatusRequest, opts ...grpc.CallOption) (*GetJobStatusResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetJobStatusResponse)
	err := c.cc.Invoke(ctx, JobService_GetJobStatus_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *jobServiceClient) RegisterAgent(ctx context.Context, in *RegisterAgentRequest, opts ...grpc.CallOption) (*RegisterAgentResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(RegisterAgentResponse)
	err := c.cc.Invoke(ctx, JobService_RegisterAgent_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *jobServiceClient) UpdateJobStatus(ctx context.Context, in *UpdateJobStatusRequest, opts ...grpc.CallOption) (*UpdateJobStatusResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(UpdateJobStatusResponse)
	err := c.cc.Invoke(ctx, JobService_UpdateJobStatus_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *jobServiceClient) FetchJob(ctx context.Context, in *FetchJobRequest, opts ...grpc.CallOption) (*FetchJobResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(FetchJobResponse)
	err := c.cc.Invoke(ctx, JobService_FetchJob_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// JobServiceServer is the server API for JobService service.
// All implementations must embed UnimplementedJobServiceServer
// for forward compatibility
//
// JobService is the main service for managing jobs.
type JobServiceServer interface {
	// SubmitJob submits a new test job.
	SubmitJob(context.Context, *SubmitJobRequest) (*SubmitJobResponse, error)
	// GetJobStatus retrieves the status of a job.
	GetJobStatus(context.Context, *GetJobStatusRequest) (*GetJobStatusResponse, error)
	// RegisterAgent allows an agent to register with the orchestrator.
	RegisterAgent(context.Context, *RegisterAgentRequest) (*RegisterAgentResponse, error)
	// UpdateJobStatus is used by an agent to report job progress.
	UpdateJobStatus(context.Context, *UpdateJobStatusRequest) (*UpdateJobStatusResponse, error)
	// FetchJob fetches a job for a given target capability.
	FetchJob(context.Context, *FetchJobRequest) (*FetchJobResponse, error)
	mustEmbedUnimplementedJobServiceServer()
}

// UnimplementedJobServiceServer must be embedded to have forward compatible implementations.
type UnimplementedJobServiceServer struct {
}

func (UnimplementedJobServiceServer) SubmitJob(context.Context, *SubmitJobRequest) (*SubmitJobResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitJob not implemented")
}
func (UnimplementedJobServiceServer) GetJobStatus(context.Context, *GetJobStatusRequest) (*GetJobStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetJobStatus not implemented")
}
func (UnimplementedJobServiceServer) RegisterAgent(context.Context, *RegisterAgentRequest) (*RegisterAgentResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegisterAgent not implemented")
}
func (UnimplementedJobServiceServer) UpdateJobStatus(context.Context, *UpdateJobStatusRequest) (*UpdateJobStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateJobStatus not implemented")
}
func (UnimplementedJobServiceServer) FetchJob(context.Context, *FetchJobRequest) (*FetchJobResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FetchJob not implemented")
}
func (UnimplementedJobServiceServer) mustEmbedUnimplementedJobServiceServer() {}

// UnsafeJobServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to JobServiceServer will
// result in compilation errors.
type UnsafeJobServiceServer interface {
	mustEmbedUnimplementedJobServiceServer()
}

func RegisterJobServiceServer(s grpc.ServiceRegistrar, srv JobServiceServer) {
	s.RegisterService(&JobService_ServiceDesc, srv)
}

func _JobService_SubmitJob_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SubmitJobRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(JobServiceServer).SubmitJob(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: JobService_SubmitJob_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(JobServiceServer).SubmitJob(ctx, req.(*SubmitJobRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _JobService_GetJobStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetJobStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(JobServiceServer).GetJobStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: JobService_GetJobStatus_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(JobServiceServer).GetJobStatus(ctx, req.(*GetJobStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _JobService_RegisterAgent_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterAgentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(JobServiceServer).RegisterAgent(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: JobService_RegisterAgent_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(JobServiceServer).RegisterAgent(ctx, req.(*RegisterAgentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _JobService_UpdateJobStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateJobStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(JobServiceServer).UpdateJobStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: JobService_UpdateJobStatus_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(JobServiceServer).UpdateJobStatus(ctx, req.(*UpdateJobStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _JobService_FetchJob_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FetchJobRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(JobServiceServer).FetchJob(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: JobService_FetchJob_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(JobServiceServer).FetchJob(ctx, req.(*FetchJobRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// JobService_ServiceDesc is the grpc.ServiceDesc for JobService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var JobService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "job_service.JobService",
	HandlerType: (*JobServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SubmitJob",
			Handler:    _JobService_SubmitJob_Handler,
		},
		{
			MethodName: "GetJobStatus",
			Handler:    _JobService_GetJobStatus_Handler,
		},
		{
			MethodName: "RegisterAgent",
			Handler:    _JobService_RegisterAgent_Handler,
		},
		{
			MethodName: "UpdateJobStatus",
			Handler:    _JobService_UpdateJobStatus_Handler,
		},
		{
			MethodName: "FetchJob",
			Handler:    _JobService_FetchJob_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/proto/job_service.proto",
}
