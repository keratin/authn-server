package services

import (
	"bytes"
	"image/png"

	"github.com/keratin/authn-server/app/data"
	"github.com/keratin/authn-server/lib/route"
	"github.com/pkg/errors"
	"github.com/pquerna/otp/totp"
)

//TOTPCreator handles the creation and storage of new TOTP tokens
func TOTPCreator(accountStore data.AccountStore, totpCache data.TOTPCache, accountID int, audience *route.Domain) (*bytes.Buffer, error) {

	account, err := AccountGetter(accountStore, accountID)
	if err != nil {
		return nil, err
	}

	//Generate totp key
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      audience.Hostname,
		AccountName: account.Username,
	})
	if err != nil {
		return nil, err
	}

	if err := totpCache.CacheTOTPSecret(account.ID, []byte(key.Secret())); err != nil {
		return nil, errors.Wrap(err, "TOTPCreator")
	}

	//Convert to png
	var buf bytes.Buffer
	img, err := key.Image(200, 200)
	if err != nil {
		return nil, err
	}
	png.Encode(&buf, img)

	return &buf, nil
}
