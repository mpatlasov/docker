package graphdriver

import (
	"fmt"
	"os/exec"
	"strings"
)

type ProxyAPI struct {
	driver Driver
	Root   string
	CtName string
	mounts map[string]string
}

////////////// Init ////////////////

type InitArgs struct {
	DriverName string
	Home       string
	Options    []string
}

type InitReply struct{}

func (p *ProxyAPI) Init(args *InitArgs, reply *InitReply) error {
	var err error
	p.mounts = make(map[string]string)

	p.driver, err = GetDriver(args.DriverName, p.Root+args.Home, args.Options)
	*reply = InitReply{}
	return err
}

////////////// Status ////////////////

type StatusArgs struct{}

type StatusReply struct {
	Status [][2]string
}

func (p *ProxyAPI) Status(args *StatusArgs, reply *StatusReply) error {
	if p.driver == nil {
		return fmt.Errorf("driver not initialized")
	}

	*reply = StatusReply{p.driver.Status()}
	return nil
}

////////////// Create ////////////////

type CreateArgs struct {
	Id, Parent string
}

type CreateReply struct{}

func (p *ProxyAPI) Create(args *CreateArgs, reply *CreateReply) error {
	if p.driver == nil {
		return fmt.Errorf("driver not initialized")
	}

	*reply = CreateReply{}
	return p.driver.Create(args.Id, args.Parent)
}

////////////// Remove ////////////////

type RemoveArgs struct {
	Id string
}

type RemoveReply struct{}

func (p *ProxyAPI) Remove(args *RemoveArgs, reply *RemoveReply) error {
	if p.driver == nil {
		return fmt.Errorf("driver not initialized")
	}

	*reply = RemoveReply{}
	return p.driver.Remove(args.Id)
}

////////////// Get ////////////////

type GetArgs struct {
	Id, MountLabel string
}

type GetReply struct {
	Dir string
}

func (p *ProxyAPI) Get(args *GetArgs, reply *GetReply) error {
	if p.driver == nil {
		return fmt.Errorf("driver not initialized")
	}

	mp, err := p.driver.Get(args.Id, args.MountLabel)
	if err != nil {
		return err
	}

	if !strings.HasPrefix(mp, p.Root) {
		p.driver.Put(args.Id)
		return fmt.Errorf("Get(%s) returned path=%s without prefix=%s", args.Id, mp, p.Root)
	}

	cli_dir := strings.TrimPrefix(mp, p.Root)
	p.mounts[args.Id] = cli_dir

	// Here, instead of calling external program, we have to ask docker
	// daemon running on host system to mount cli_dir into the namespace
	// of p.CtName container (setns(2)+mount(2))
	err = exec.Command("dock-mount", p.CtName, p.Root, cli_dir).Run()
	if err != nil {
		fmt.Println("ProxyAPI.Get dock-mount error:", err)
		p.driver.Put(args.Id)
		return err
	}
	*reply = GetReply{cli_dir}

	return nil
}

////////////// Put ////////////////

type PutArgs struct {
	Id string
}

type PutReply struct{}

func (p *ProxyAPI) Put(args *PutArgs, reply *PutReply) error {
	if p.driver == nil {
		return fmt.Errorf("driver not initialized")
	}

	if _, ok := p.mounts[args.Id]; !ok {
		return fmt.Errorf("ProxyAPI.Put mounts[%s] not exists", args.Id)
	}

	// Here, instead of calling external program, we have to ask docker
	// daemon running on host system to umount p.mounts[args.Id] from
	// the namespace of p.CtName container (setns(2)+umount(2))
	err := exec.Command("dock-umount", p.CtName, p.mounts[args.Id]).Run()
	if err != nil {
		fmt.Println("ProxyAPI.Put dock-umount error:", err)
		return err
	}
	delete(p.mounts, args.Id)

	p.driver.Put(args.Id)

	*reply = PutReply{}
	return nil
}

////////////// Exists ////////////////

type ExistsArgs struct {
	Id string
}

type ExistsReply struct {
	Exists bool
}

func (p *ProxyAPI) Exists(args *ExistsArgs, reply *ExistsReply) error {
	if p.driver == nil {
		return fmt.Errorf("driver not initialized")
	}

	*reply = ExistsReply{p.driver.Exists(args.Id)}
	return nil
}

////////////// Cleanup ////////////////

type CleanupArgs struct{}

type CleanupReply struct{}

func (p *ProxyAPI) Cleanup(args *CleanupArgs, reply *CleanupReply) error {
	if p.driver == nil {
		return fmt.Errorf("driver not initialized")
	}

	*reply = CleanupReply{}
	return p.driver.Cleanup()
}

////////////// GetMetadata ////////////////

type GetMetadataArgs struct {
	Id string
}

type GetMetadataReply struct {
	MInfo map[string]string
}

func (p *ProxyAPI) GetMetadata(args *GetMetadataArgs, reply *GetMetadataReply) error {
	if p.driver == nil {
		return fmt.Errorf("driver not initialized")
	}

	minfo, err := p.driver.GetMetadata(args.Id)
	if err != nil {
		return err
	}

	*reply = GetMetadataReply{minfo}
	return nil
}
