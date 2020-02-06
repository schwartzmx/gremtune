
# This target (taken from: https://gist.github.com/prwhite/8168133) is an easy way to print out a usage/ help of all make targets.
# For all make targets the text after \#\# will be printed.
help: ## Prints the help
	@echo "$$(grep -hE '^\S+:.*##' $(MAKEFILE_LIST) | sed -e 's/:.*##\s*/:/' -e 's/^\(.\+\):\(.*\)/\1\:\2/' | column -c2 -t -s :)"

up: sep ## Start infra components
	@echo "--> Startup (+build) infra components"
	@docker-compose up --build -d

down: sep ## Stop infra components
	@echo "--> Stop infra components"
	@docker-compose down --remove-orph

sep:
	@echo "----------------------------------------------------------------------------------"