package sms

import (
	"context"

	"github.com/gidyon/services/pkg/api/messaging/sms"
	"github.com/gidyon/services/pkg/auth"
	"github.com/gidyon/services/pkg/utils/errs"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/grpclog"
)

type sendSMSFunc func(*Options, *sms.SMS) error

type smsAPIServer struct {
	logger  grpclog.LoggerV2
	authAPI auth.Interface
	sendSMS sendSMSFunc
	opt     *Options
}

// Options contains parameters passed while calling NewSMSAPIServer
type Options struct {
	Logger        grpclog.LoggerV2
	JWTSigningKey []byte
	APIKey        string
	AuthToken     string
	APIUsername   string
	APIPassword   string
	APIURL        string
}

// NewSMSAPIServer creates a new sms API server
func NewSMSAPIServer(ctx context.Context, opt *Options) (sms.SMSAPIServer, error) {
	// Validation
	var err error
	switch {
	case ctx == nil:
		err = errs.NilObject("context")
	case opt == nil:
		err = errs.NilObject("options")
	case opt.Logger == nil:
		err = errs.NilObject("logger")
	case opt.JWTSigningKey == nil:
		err = errs.NilObject("jwt key")
	case opt.APIKey == "":
		err = errs.MissingField("api key")
	case opt.AuthToken == "":
		err = errs.MissingField("auth token")
	case opt.APIUsername == "":
		err = errs.MissingField("api username")
	case opt.APIPassword == "":
		err = errs.MissingField("api password")
	case opt.APIURL == "":
		err = errs.MissingField("api url")
	}
	if err != nil {
		return nil, err
	}

	// Auth API
	authAPI, err := auth.NewAPI(opt.JWTSigningKey, "SMS API", "users")
	if err != nil {
		return nil, err
	}

	// API server
	smsAPI := &smsAPIServer{
		logger:  opt.Logger,
		authAPI: authAPI,
		sendSMS: sendSmsAT, // uses Africa's talinkg implementation
		opt:     opt,
	}

	if smsAPI.sendSMS == nil {
		return nil, errs.NilObject("sender function")
	}

	return smsAPI, nil
}

func (api *smsAPIServer) SendSMS(
	ctx context.Context, sendReq *sms.SMS,
) (*empty.Empty, error) {
	// Request must not be nil
	if sendReq == nil {
		return nil, errs.NilObject("SMS")
	}

	// Authenticate request
	err := api.authAPI.AuthenticateRequest(ctx)
	if err != nil {
		return nil, err
	}

	// Validate sms
	err = validateSMS(sendReq)
	if err != nil {
		return nil, err
	}

	// Send sms
	go func() {
		err = api.sendSMS(api.opt, sendReq)
		if err != nil {
			api.logger.Errorf("failed to send sms: %v", err)
		}
	}()

	return &empty.Empty{}, nil
}

func validateSMS(smsPB *sms.SMS) error {
	var err error
	switch {
	case len(smsPB.DestinationPhones) == 0:
		err = errs.MissingField("destination phones")
	case smsPB.Keyword == "":
		err = errs.MissingField("keyword")
	case smsPB.Message == "":
		err = errs.MissingField("message")
	}
	return err
}
