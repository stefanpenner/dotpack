.PHONY: build build-darwin push status install clean

build:
	./dotpack build

build-darwin:
	./dotpack build --os darwin

push:
	./dotpack push $(NAS_HOST)

status:
	./dotpack status $(NAS_HOST)

install:
	./dotpack install

clean:
	./dotpack clean
