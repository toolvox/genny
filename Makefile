.PHONY: make
make:
	@echo "Building..."
	@go build ./cmd/genny
	@echo "Installing..."
	@go install ./cmd/genny
	@echo "Done!"

copy:
	@mkdir -p ./genny
	@cp -r ../../_shloof/portfolio/genny/* ./shloof
