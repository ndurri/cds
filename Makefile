.PHONY: all clean
all: auth-redirect.zip authorize.zip detect.zip mail.zip reply.zip request.zip response.zip submit.zip token.zip lib.zip
%.zip:
	cd $(basename $@) && go mod tidy && go build -ldflags="-w -s" -o bootstrap && zip ../$@ bootstrap
lib/lambda.so:
	cd lib/lambda && go mod tidy && go build -buildmode=plugin -ldflags="-w -s" -o ../lambda.so
lib/aws.so:
	cd lib/aws && go mod tidy && go build -buildmode=plugin -ldflags="-w -s" -o ../aws.so
lib.zip: lib/lambda.so lib/aws.so
	zip lib.zip lib/*.so
clean:
	rm *.zip */bootstrap lib/*.so