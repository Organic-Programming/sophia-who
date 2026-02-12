package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"

	pb "sophia-who/proto"

	"google.golang.org/grpc"
)

const defaultPort = "50051"

// server implements the SophiaWhoService gRPC interface.
// It reuses the same logic as the CLI commands.
type server struct {
	pb.UnimplementedSophiaWhoServiceServer
}

func (s *server) CreateIdentity(ctx context.Context, req *pb.CreateIdentityRequest) (*pb.CreateIdentityResponse, error) {
	id := NewIdentity()

	if req.GivenName == "" || req.FamilyName == "" || req.Motto == "" || req.Composer == "" {
		return nil, fmt.Errorf("given_name, family_name, motto, and composer are required")
	}

	id.GivenName = req.GivenName
	id.FamilyName = req.FamilyName
	id.Motto = req.Motto
	id.Composer = req.Composer
	id.Clade = cladeToString(req.Clade)
	id.Reproduction = reproductionToString(req.Reproduction)

	if req.Lang != "" {
		id.Lang = req.Lang
	}
	if len(req.Aliases) > 0 {
		id.Aliases = req.Aliases
	}
	if req.WrappedLicense != "" {
		id.WrappedLicense = req.WrappedLicense
	}

	// Determine output directory
	outputDir := req.OutputDir
	if outputDir == "" {
		dirName := strings.ToLower(id.GivenName + "-" + strings.TrimSuffix(id.FamilyName, "?"))
		dirName = strings.ReplaceAll(dirName, " ", "-")
		outputDir = filepath.Join(".holon", dirName)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("cannot create directory: %w", err)
	}

	outputPath := filepath.Join(outputDir, "HOLON.md")
	if err := writeHolonMD(id, outputPath); err != nil {
		return nil, err
	}

	return &pb.CreateIdentityResponse{
		Identity: identityToProto(id),
		FilePath: outputPath,
	}, nil
}

func (s *server) ShowIdentity(ctx context.Context, req *pb.ShowIdentityRequest) (*pb.ShowIdentityResponse, error) {
	path, err := findHolonByUUID(req.Uuid)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", path, err)
	}

	id, _, err := parseFrontmatter(data)
	if err != nil {
		return nil, err
	}

	return &pb.ShowIdentityResponse{
		Identity:   identityToProto(id),
		FilePath:   path,
		RawContent: string(data),
	}, nil
}

func (s *server) ListIdentities(ctx context.Context, req *pb.ListIdentitiesRequest) (*pb.ListIdentitiesResponse, error) {
	holons, err := findAllHolons()
	if err != nil {
		return nil, err
	}

	var pbHolons []*pb.HolonIdentity
	for _, h := range holons {
		pbHolons = append(pbHolons, identityToProto(h))
	}

	return &pb.ListIdentitiesResponse{
		Identities: pbHolons,
	}, nil
}

func (s *server) PinVersion(ctx context.Context, req *pb.PinVersionRequest) (*pb.PinVersionResponse, error) {
	path, err := findHolonByUUID(req.Uuid)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", path, err)
	}

	id, _, err := parseFrontmatter(data)
	if err != nil {
		return nil, err
	}

	// Update pinning fields (only non-empty values)
	if req.BinaryPath != "" {
		id.BinaryPath = req.BinaryPath
	}
	if req.BinaryVersion != "" {
		id.BinaryVersion = req.BinaryVersion
	}
	if req.GitTag != "" {
		id.GitTag = req.GitTag
	}
	if req.GitCommit != "" {
		id.GitCommit = req.GitCommit
	}
	if req.Os != "" {
		id.OS = req.Os
	}
	if req.Arch != "" {
		id.Arch = req.Arch
	}

	if err := writeHolonMD(id, path); err != nil {
		return nil, err
	}

	return &pb.PinVersionResponse{
		Identity: identityToProto(id),
	}, nil
}

// runServe starts the gRPC server.
func runServe() error {
	port := defaultPort
	if len(os.Args) > 2 {
		port = os.Args[2]
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", port, err)
	}

	s := grpc.NewServer()
	pb.RegisterSophiaWhoServiceServer(s, &server{})

	log.Printf("Sophia Who? gRPC server listening on :%s", port)
	return s.Serve(lis)
}

// --- Conversion helpers ---

func identityToProto(id Identity) *pb.HolonIdentity {
	return &pb.HolonIdentity{
		Uuid:           id.UUID,
		GivenName:      id.GivenName,
		FamilyName:     id.FamilyName,
		Motto:          id.Motto,
		Composer:       id.Composer,
		Clade:          stringToClade(id.Clade),
		Status:         stringToStatus(id.Status),
		Born:           id.Born,
		Parents:        id.Parents,
		Reproduction:   stringToReproduction(id.Reproduction),
		BinaryPath:     id.BinaryPath,
		BinaryVersion:  id.BinaryVersion,
		GitTag:         id.GitTag,
		GitCommit:      id.GitCommit,
		Os:             id.OS,
		Arch:           id.Arch,
		Dependencies:   id.Dependencies,
		Aliases:        id.Aliases,
		WrappedLicense: id.WrappedLicense,
		GeneratedBy:    id.GeneratedBy,
		Lang:           id.Lang,
		ProtoStatus:    stringToStatus(id.ProtoStatus),
	}
}

func cladeToString(c pb.Clade) string {
	m := map[pb.Clade]string{
		pb.Clade_DETERMINISTIC_PURE:       "deterministic/pure",
		pb.Clade_DETERMINISTIC_STATEFUL:   "deterministic/stateful",
		pb.Clade_DETERMINISTIC_IO_BOUND:   "deterministic/io_bound",
		pb.Clade_PROBABILISTIC_GENERATIVE: "probabilistic/generative",
		pb.Clade_PROBABILISTIC_PERCEPTUAL: "probabilistic/perceptual",
		pb.Clade_PROBABILISTIC_ADAPTIVE:   "probabilistic/adaptive",
	}
	if s, ok := m[c]; ok {
		return s
	}
	return "deterministic/pure"
}

func stringToClade(s string) pb.Clade {
	m := map[string]pb.Clade{
		"deterministic/pure":       pb.Clade_DETERMINISTIC_PURE,
		"deterministic/stateful":   pb.Clade_DETERMINISTIC_STATEFUL,
		"deterministic/io_bound":   pb.Clade_DETERMINISTIC_IO_BOUND,
		"probabilistic/generative": pb.Clade_PROBABILISTIC_GENERATIVE,
		"probabilistic/perceptual": pb.Clade_PROBABILISTIC_PERCEPTUAL,
		"probabilistic/adaptive":   pb.Clade_PROBABILISTIC_ADAPTIVE,
	}
	if c, ok := m[s]; ok {
		return c
	}
	return pb.Clade_CLADE_UNSPECIFIED
}

func stringToStatus(s string) pb.Status {
	m := map[string]pb.Status{
		"draft":      pb.Status_DRAFT,
		"stable":     pb.Status_STABLE,
		"deprecated": pb.Status_DEPRECATED,
		"dead":       pb.Status_DEAD,
	}
	if st, ok := m[s]; ok {
		return st
	}
	return pb.Status_STATUS_UNSPECIFIED
}

func reproductionToString(r pb.ReproductionMode) string {
	m := map[pb.ReproductionMode]string{
		pb.ReproductionMode_MANUAL:      "manual",
		pb.ReproductionMode_ASSISTED:    "assisted",
		pb.ReproductionMode_AUTOMATIC:   "automatic",
		pb.ReproductionMode_AUTOPOIETIC: "autopoietic",
		pb.ReproductionMode_BRED:        "bred",
	}
	if s, ok := m[r]; ok {
		return s
	}
	return "manual"
}

func stringToReproduction(s string) pb.ReproductionMode {
	m := map[string]pb.ReproductionMode{
		"manual":      pb.ReproductionMode_MANUAL,
		"assisted":    pb.ReproductionMode_ASSISTED,
		"automatic":   pb.ReproductionMode_AUTOMATIC,
		"autopoietic": pb.ReproductionMode_AUTOPOIETIC,
		"bred":        pb.ReproductionMode_BRED,
	}
	if r, ok := m[s]; ok {
		return r
	}
	return pb.ReproductionMode_REPRODUCTION_UNSPECIFIED
}
