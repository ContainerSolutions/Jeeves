module github.com/ContainerSolutions/jeeves

go 1.15

replace k8s.io/client-go => k8s.io/client-go v0.18.12

require (
	cloud.google.com/go v0.38.0 // indirect
	github.com/Azure/go-autorest/autorest v0.9.0 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/google/btree v1.0.0 // indirect
	github.com/google/go-github/v32 v32.1.0
	github.com/gophercloud/gophercloud v0.1.0 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/iovisor/kubectl-trace v0.1.0-rc.1
	github.com/sirupsen/logrus v1.7.0
	github.com/slack-go/slack v0.7.3
	github.com/stretchr/testify v1.4.0
	golang.org/x/crypto v0.0.0-20200220183623-bac4c82f6975 // indirect
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4 // indirect
	k8s.io/api v0.18.12
	k8s.io/apimachinery v0.18.12
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/utils v0.0.0-20200324210504-a9aa75ae1b89 // indirect
)
