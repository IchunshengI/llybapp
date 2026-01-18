package login

import pb "llyb-backend/proto"

func Handle(req *pb.LoginRequest) (bool, string) {
	_ = req
	return true, "登录成功"
}
