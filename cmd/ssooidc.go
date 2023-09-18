package cmd

//const (
//	grantType  string = "urn:ietf:params:oauth:grant-type:device_code"
//	clientType string = "public"
//	clientName string = "aws-sso-util"
//)
//
//// ClientInformation is used to store client information
//type ClientInformation struct {
//	AccessTokenExpiresAt    time.Time
//	AccessToken             string
//	ClientID                string
//	ClientSecret            string
//	ClientSecretExpiresAt   string
//	DeviceCode              string
//	VerificationURIComplete string
//	StartURL                string
//}
//
//// IsExpired is used to tell if AccessToken is expired in client information
//func (ati ClientInformation) IsExpired() bool {
//	if ati.AccessTokenExpiresAt.Before(time.Now()) {
//		return true
//	}
//	return false
//}
//
//// OIDCInformation contains common info for sso oidc
//type OIDCInformation struct {
//	Client *ssooidc.Client
//	URL    string
//}
//
//// ProcessClientInformation tries to read available ClientInformation.
//// If no ClientInformation is available it retrieves and creates new one and writes this information to disk
//// If the start url is overridden via flag and differs from the previous one, a new Client is registered for the given start url.
//// When the ClientInformation.AccessToken is expired, it starts retrieving a new AccessToken
//func (o OIDCInformation) ProcessClientInformation() (ClientInformation, error) {
//	clientInformation, err := ReadClientInformation(ClientInfoFileDestination())
//	if err != nil || clientInformation.StartURL != o.URL {
//		var clientInfoPointer *ClientInformation
//		clientInfoPointer = o.registerClient()
//		clientInfoPointer = retrieveToken(o.Client, clientInfoPointer)
//		WriteStructToFile(clientInfoPointer, ClientInfoFileDestination())
//		clientInformation = *clientInfoPointer
//	} else if clientInformation.IsExpired() {
//		log.Println("AccessToken expired. Start retrieving a new AccessToken.")
//		clientInformation = o.handleOutdatedAccessToken(clientInformation)
//	}
//	return clientInformation, err
//}
//
//// handleOutdatedAccessToken handles client information if AccessToken is expired
//func (o OIDCInformation) handleOutdatedAccessToken(clientInformation ClientInformation) ClientInformation {
//	registerClientOutput := ssooidc.RegisterClientOutput{ClientId: &clientInformation.ClientID, ClientSecret: &clientInformation.ClientSecret}
//	deviceAuth, err := o.startDeviceAuthorization(&registerClientOutput)
//	if err != nil {
//		log.Println("Failed to authorize device. Regenerating AccessToken")
//		var clientInfoPointer *ClientInformation
//		clientInfoPointer = o.registerClient()
//		clientInfoPointer = retrieveToken(o.Client, clientInfoPointer)
//		WriteStructToFile(clientInfoPointer, ClientInfoFileDestination())
//		return *clientInfoPointer
//	}
//
//	clientInformation.DeviceCode = *deviceAuth.DeviceCode
//	var clientInfoPointer *ClientInformation
//	clientInfoPointer = retrieveToken(o.Client, &clientInformation)
//	WriteStructToFile(clientInfoPointer, ClientInfoFileDestination())
//	return *clientInfoPointer
//}
//
//// RegisterClient is used to start device auth
//func (o OIDCInformation) registerClient() *ClientInformation {
//	cn := clientName
//	ct := clientType
//
//	input := ssooidc.RegisterClientInput{ClientName: &cn, ClientType: &ct}
//	output, err := o.Client.RegisterClient(ctx, &input)
//	check(err)
//
//	deviceAuth, err := o.startDeviceAuthorization(output)
//	check(err)
//
//	return &ClientInformation{
//		ClientID:                *output.ClientId,
//		ClientSecret:            *output.ClientSecret,
//		ClientSecretExpiresAt:   strconv.FormatInt(output.ClientSecretExpiresAt, 10),
//		DeviceCode:              *deviceAuth.DeviceCode,
//		VerificationURIComplete: *deviceAuth.VerificationUriComplete,
//		StartURL:                o.URL,
//	}
//}
//
//func (o OIDCInformation) startDeviceAuthorization(rco *ssooidc.RegisterClientOutput) (ssooidc.StartDeviceAuthorizationOutput, error) {
//	output, err := o.Client.StartDeviceAuthorization(ctx, &ssooidc.StartDeviceAuthorizationInput{
//		ClientId:     rco.ClientId,
//		ClientSecret: rco.ClientSecret,
//		StartUrl:     &o.URL,
//	})
//	if err != nil {
//		return ssooidc.StartDeviceAuthorizationOutput{}, fmt.Errorf("Encountered error at startDeviceAuthorization: %w", err)
//	}
//	log.Println("Please verify your client request: " + *output.VerificationUriComplete)
//	openURLInBrowser(*output.VerificationUriComplete)
//	return *output, nil
//}
//
//// open browser for supported runtimes
//func openURLInBrowser(url string) {
//	var err error
//
//	switch runtime.GOOS {
//	case "linux":
//		err = exec.Command("xdg-open", url).Start()
//	case "darwin":
//		err = exec.Command("open", url).Start()
//	case "windows":
//		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
//	default:
//		err = fmt.Errorf("could not open %s - unsupported platform. Please open the URL manually", url)
//	}
//	if err != nil {
//		log.Fatal(err)
//	}
//
//}
//
//// used to create a CreateTokenInput
//func generateCreateTokenInput(clientInformation *ClientInformation) ssooidc.CreateTokenInput {
//	gtp := grantType
//	return ssooidc.CreateTokenInput{
//		ClientId:     &clientInformation.ClientID,
//		ClientSecret: &clientInformation.ClientSecret,
//		DeviceCode:   &clientInformation.DeviceCode,
//		GrantType:    &gtp,
//	}
//}
//
//// retrieveToken is used to create the access token from the sso session
//// this is obtained after auth in the browser.
//func retrieveToken(client *ssooidc.Client, info *ClientInformation) *ClientInformation {
//	input := generateCreateTokenInput(info)
//	// need loop to prevent errors while waiting on auth through browser
//	for {
//		cto, err := client.CreateToken(ctx, &input)
//		if err != nil {
//			var ae smithy.APIError
//			if errors.As(err, &ae) {
//				if ae.ErrorCode() == "AuthorizationPendingException" {
//					log.Println("Waiting on authorization..")
//					time.Sleep(5 * time.Second)
//					continue
//				}
//				log.Fatalf("Encountered an error while retrieveToken: %v", err)
//			}
//		}
//		info.AccessToken = *cto.AccessToken
//		info.AccessTokenExpiresAt = time.Now().Add(time.Hour * 8)
//		//info.AccessTokenExpiresAt = timer.Now().Add(time.Hour*8 - time.Minute*5)
//		return info
//	}
//}
