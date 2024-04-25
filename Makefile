.PHONY: all clean
all: auth-redirect.zip authorize.zip detect.zip mail.zip reply.zip request.zip response.zip submit.zip token.zip
%.zip:
	cd $(basename $@) && go mod tidy && go build -o bootstrap && zip ../$@ bootstrap
clean:
	rm *.zip */bootstrap