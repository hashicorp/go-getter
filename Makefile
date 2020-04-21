start-smb:
	docker-compose up -d samba

smbtests-prepare:
	@docker-compose up -d
	@sleep 60
	@docker cp ./ gogetter:/go-getter/
	@docker exec -it samba bash -c "echo 'Hello' > data/file.txt && mkdir -p data/subdir  && echo 'Hello' > data/subdir/file.txt"
	@docker exec -it gogetter bash -c "go mod download && apt-get update && apt-get -y install smbclient"


smbtests:
	@docker cp ./ gogetter:/go-getter/
	@docker exec -it gogetter bash -c "env ACC_SMB_TEST=1 go test -v ./... -run=TestSmbGetter_"