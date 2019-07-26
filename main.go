package main

import (
	"context"
	"log"
	"sync"

	"github.com/micro/go-micro"

	pb "github.com/finalsatan/shiiip-consignment/proto/consignment"
	vesselPb "github.com/finalsatan/shiiip-vessel/proto/vessel"
)

type repository interface {
	Create(*pb.Consignment) (*pb.Consignment, error)
	GetAll() []*pb.Consignment
}

// Repository - Dummy repository, this simulates the use of a datastore
// of some kind. We'll replace this with a real implementation later on.
type Repository struct {
	mu           sync.Mutex
	consignments []*pb.Consignment
}

// Create a new consignment
func (repo *Repository) Create(consignment *pb.Consignment) (*pb.Consignment, error) {
	repo.mu.Lock()
	updated := append(repo.consignments, consignment)
	repo.consignments = updated
	repo.mu.Unlock()
	return consignment, nil
}

func (repo *Repository) GetAll() []*pb.Consignment {
	return repo.consignments
}

// Service should implement all of the methods to satisfy the service
// we defined in our protobuf definition. You can check the interface
// in the generated code itself for the exact signature etc to give
// you a better idea.
type service struct {
	repo repository
	vesselClient vesselPb.VesselServiceClient
}

// CreateConsignment -  we created just one method on our service,
// which is a create method, which takes a context and a request as an
// argument, these are handled by the gRPC server.
func (s *service) CreateConsignment(ctx context.Context, req *pb.Consignment, resp *pb.Response) error {
	vesselResp,err := s.vesselClient.FindAvailable(context.Background(),&vesselPb.Specification{
		MaxWeight: req.Weight,
		Capacity:int32(len(req.Containers)),
	})
	if err != nil {
		return err
	}

	req.VesselId = vesselResp.Vessel.Id

	// Save our consignment
	consignment, err := s.repo.Create(req)
	if err != nil {
		return err
	}

	// Return matching the `Response` message we created in our protobuf definition.
	resp.Created = true
	resp.Consignment = consignment

	return nil
}

func (s *service) GetConsignments(ctx context.Context, req *pb.GetRequest, resp *pb.Response) error {
	resp.Consignments = s.repo.GetAll()
	return nil
}

func main() {
	repo := &Repository{}

	// Create a new service. Optionally include some options here.
	srv := micro.NewService(
		micro.Name("shiiip.consignment"),
	)

	// Init will parse the command line flags.
	srv.Init()

	vesselClient := vesselPb.NewVesselServiceClient("shiiip.vessel",srv.Client())

	// Register handler
	pb.RegisterShippingServiceHandler(srv.Server(), &service{repo,vesselClient})

	// Run the server
	if err := srv.Run(); err != nil {
		log.Fatalf("Failed to run consignment service server: %v", err)
	}
}
