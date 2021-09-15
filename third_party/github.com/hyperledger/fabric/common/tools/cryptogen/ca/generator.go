/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package ca

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"github.com/hyperledger/fabric/bccsp/cncc"
	"github.com/hyperledger/fabric/bccsp/signer"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric/bccsp/gm"
	"github.com/hyperledger/fabric/bccsp/utils"
	"github.com/hyperledger/fabric/common/tools/cryptogen/csp"
	"github.com/tjfoc/gmsm/sm2"
)

type CA struct {
	Name               string
	Country            string
	Province           string
	Locality           string
	OrganizationalUnit string
	StreetAddress      string
	PostalCode         string
	//SignKey  *ecdsa.PrivateKey
	Signer      crypto.Signer
	SignCert    *x509.Certificate
	SignSm2Cert *sm2.Certificate
	Sm2Key      bccsp.Key
}

// NewCA creates an instance of CA and saves the signing key pair in
// baseDir/name
func GetCA(baseDir string) (*CA, error) {

	var ca *CA

	certBytes, _ := ioutil.ReadFile("/root/桌面/ca.pem")

	x509Cert, err := sm2.ReadCertificateFromMem(certBytes)

	if err == nil {

		priv, signer, _ := csp.GetPrivateKey(baseDir, x509Cert.SubjectKeyId)

		if err == nil {
			ca = &CA{
				Name:               "ccpc.cn",
				Signer:             signer,
				Country:            "CN",
				Province:           "BEIJING",
				Locality:           "BEIJING",
				OrganizationalUnit: "",
				StreetAddress:      "",
				PostalCode:         "",
				SignSm2Cert:        x509Cert,
				Sm2Key:             priv,
			}
		}

	}
	return ca, nil
}

// NewCA creates an instance of CA and saves the signing key pair in
// baseDir/name
func NewCA(baseDir, org, name, country, province, locality, orgUnit, streetAddress, postalCode string) (*CA, error) {

	var response error
	var ca *CA

	err := os.MkdirAll(baseDir, 0755)
	if err == nil {
		_, signerRoot, err := csp.GenerateRootPrivateKey(baseDir)
		priv, signer, err := csp.GeneratePrivateKey(baseDir)
		response = err
		if err == nil {
			// get public signing certificate
			//ecPubKey, err := csp.GetECPublicKey(priv)
			sm2PubKey, err := csp.GetSM2PublicKey(priv)
			response = err
			if err == nil {
				template := x509Template()
				//this is a CA
				template.IsCA = true
				template.KeyUsage |= x509.KeyUsageDigitalSignature |
					x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign |
					x509.KeyUsageCRLSign
				template.ExtKeyUsage = []x509.ExtKeyUsage{
					x509.ExtKeyUsageClientAuth,
					x509.ExtKeyUsageServerAuth,
				}

				//set the organization for the subject
				subject := subjectTemplateAdditional(country, province, locality, orgUnit, streetAddress, postalCode)
				subject.Organization = []string{org}
				subject.CommonName = name

				template.Subject = subject
				template.SubjectKeyId = priv.SKI()

				//x509Cert, err := genCertificateECDSA(baseDir, name, &template, &template,
				//	ecPubKey, signer)
				sm2Template := gm.ParseX509Certificate2Sm2(&template)

				sm2Template.PublicKey = sm2PubKey
				sm2Template.NetWorkId = []byte("test")

				x509Cert, err := genCertificateGMSM2(baseDir, name, sm2Template, sm2Template, sm2PubKey, signerRoot, signer)

				response = err
				if err == nil {
					ca = &CA{
						Name:   name,
						Signer: signer,
						//SignCert:           x509Cert,
						Country:            country,
						Province:           province,
						Locality:           locality,
						OrganizationalUnit: orgUnit,
						StreetAddress:      streetAddress,
						PostalCode:         postalCode,
						SignSm2Cert:        x509Cert,
						Sm2Key:             priv,
					}
				}
			}
		}
	}
	return ca, response
}

// NewCA creates an instance of CA and saves the signing key pair in
// baseDir/name
func NewTlsCA(baseDir, org, name, country, province, locality, orgUnit, streetAddress, postalCode string) (*CA, error) {

	var response error
	var ca *CA

	err := os.MkdirAll(baseDir, 0755)
	if err == nil {
		_, signerRoot, err := csp.GenerateRootPrivateKey(baseDir)
		priv, signer, err := csp.GeneratePrivateKey(baseDir)
		response = err
		if err == nil {
			// get public signing certificate
			//ecPubKey, err := csp.GetECPublicKey(priv)
			sm2PubKey, err := csp.GetSM2PublicKey(priv)
			response = err
			if err == nil {
				template := x509Template()
				//this is a CA
				template.IsCA = true
				template.KeyUsage |= x509.KeyUsageDigitalSignature |
					x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign |
					x509.KeyUsageCRLSign
				template.ExtKeyUsage = []x509.ExtKeyUsage{
					x509.ExtKeyUsageClientAuth,
					x509.ExtKeyUsageServerAuth,
				}

				//set the organization for the subject
				subject := subjectTemplateAdditional(country, province, locality, orgUnit, streetAddress, postalCode)
				subject.Organization = []string{org}
				subject.CommonName = name

				template.Subject = subject
				template.SubjectKeyId = priv.SKI()

				//x509Cert, err := genCertificateECDSA(baseDir, name, &template, &template,
				//	ecPubKey, signer)
				sm2cert := gm.ParseX509Certificate2Sm2(&template)

				sm2cert.PublicKey = sm2PubKey

				x509Cert, err := genCertificateGMSM2(baseDir, name, sm2cert, sm2cert, sm2PubKey, signerRoot, signer)

				response = err
				if err == nil {
					ca = &CA{
						Name:   name,
						Signer: signer,
						//SignCert:           x509Cert,
						Country:            country,
						Province:           province,
						Locality:           locality,
						OrganizationalUnit: orgUnit,
						StreetAddress:      streetAddress,
						PostalCode:         postalCode,
						SignSm2Cert:        x509Cert,
						Sm2Key:             priv,
					}
				}
			}
		}
	}
	return ca, response
}

// NewRootCA creates an instance of CA and saves the signing key pair in
// baseDir/name
func NewRootCA(baseDir, org, name, country, province, locality, orgUnit, streetAddress, postalCode string) (*CA,
	error) {

	var response error
	var ca *CA

	err := os.MkdirAll(baseDir, 0755)
	if err == nil {
		priv, signer, err := csp.GenerateRootPrivateKey(baseDir)
		response = err
		if err == nil {
			// get public signing certificate
			//ecPubKey, err := csp.GetECPublicKey(priv)
			sm2PubKey, err := csp.GetSM2PublicKey(priv)
			response = err
			if err == nil {
				template := x509Template()
				//this is a CA
				template.IsCA = true
				template.KeyUsage |= x509.KeyUsageDigitalSignature |
					x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign |
					x509.KeyUsageCRLSign
				template.ExtKeyUsage = []x509.ExtKeyUsage{
					x509.ExtKeyUsageClientAuth,
					x509.ExtKeyUsageServerAuth,
				}

				//set the organization for the subject
				subject := subjectTemplateAdditional(country, province, locality, orgUnit, streetAddress, postalCode)
				subject.Organization = []string{org}
				subject.CommonName = name

				template.Subject = subject
				template.SubjectKeyId = priv.SKI()

				//x509Cert, err := genCertificateECDSA(baseDir, name, &template, &template,
				//	ecPubKey, signer)
				sm2cert := gm.ParseX509Certificate2Sm2(&template)

				sm2cert.PublicKey = sm2PubKey

				x509Cert, err := genRootCertificateGMSM2(baseDir, name, sm2cert, sm2cert, sm2PubKey, signer)

				response = err
				if err == nil {
					ca = &CA{
						Name:   name,
						Signer: signer,
						//SignCert:           x509Cert,
						Country:            country,
						Province:           province,
						Locality:           locality,
						OrganizationalUnit: orgUnit,
						StreetAddress:      streetAddress,
						PostalCode:         postalCode,
						SignSm2Cert:        x509Cert,
						Sm2Key:             priv,
					}
				}
			}
		}
	}
	return ca, response
}

// SignCertificate creates a signed certificate based on a built-in template
// and saves it in baseDir/name
func (ca *CA) SignCertificate(baseDir, name string, ous, sans []string, priv bccsp.Key, ku x509.KeyUsage, eku []x509.ExtKeyUsage) (*sm2.Certificate, error) {

	template := x509Template()
	template.KeyUsage = ku
	template.ExtKeyUsage = eku
	template.SubjectKeyId = priv.SKI()

	pub, err := csp.GetSM2PublicKey(priv)
	if err != nil {
		return nil, errors.New("Get sm2 public key error !")
	}

	//set the organization for the subject
	subject := subjectTemplateAdditional(ca.Country, ca.Province, ca.Locality, ca.OrganizationalUnit, ca.StreetAddress, ca.PostalCode)
	subject.CommonName = name

	subject.OrganizationalUnit = append(subject.OrganizationalUnit, ous...)

	template.Subject = subject
	for _, san := range sans {
		// try to parse as an IP address first
		ip := net.ParseIP(san)
		if ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, san)
		}
	}

	template.PublicKey = pub
	// cert, err := genCertificateECDSA(baseDir, name, &template, ca.SignCert,
	// 	pub, ca.Signer)

	sm2Tpl := gm.ParseX509Certificate2Sm2(&template)
	cert, err := genCertificateGMSM2(baseDir, name, sm2Tpl, ca.SignSm2Cert, pub, ca.Signer, ca.Signer)

	if err != nil {
		return nil, err
	}

	return cert, nil
}

// SignCertificate creates a signed certificate based on a built-in template
// and saves it in baseDir/name
func (ca *CA) SignTlsCertificate(baseDir, name string, ous, sans []string, pub *ecdsa.PublicKey,
	ku x509.KeyUsage, eku []x509.ExtKeyUsage) (*x509.Certificate, error) {

	template := x509Template()
	template.KeyUsage = ku
	template.ExtKeyUsage = eku

	//set the organization for the subject
	subject := subjectTemplateAdditional(ca.Country, ca.Province, ca.Locality, ca.OrganizationalUnit, ca.StreetAddress, ca.PostalCode)
	subject.CommonName = name

	subject.OrganizationalUnit = append(subject.OrganizationalUnit, ous...)

	template.Subject = subject
	for _, san := range sans {
		// try to parse as an IP address first
		ip := net.ParseIP(san)
		if ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, san)
		}
	}

	cert, err := genCertificateECDSA(baseDir, name, &template, ca.SignCert,
		pub, ca.Signer)

	if err != nil {
		return nil, err
	}

	return cert, nil
}

// default template for X509 subject
func subjectTemplate() pkix.Name {
	return pkix.Name{
		Country:  []string{"US"},
		Locality: []string{"San Francisco"},
		Province: []string{"California"},
	}
}

// Additional for X509 subject
func subjectTemplateAdditional(country, province, locality, orgUnit, streetAddress, postalCode string) pkix.Name {
	name := subjectTemplate()
	if len(country) >= 1 {
		name.Country = []string{country}
	}
	if len(province) >= 1 {
		name.Province = []string{province}
	}

	if len(locality) >= 1 {
		name.Locality = []string{locality}
	}
	if len(orgUnit) >= 1 {
		name.OrganizationalUnit = []string{orgUnit}
	}
	if len(streetAddress) >= 1 {
		name.StreetAddress = []string{streetAddress}
	}
	if len(postalCode) >= 1 {
		name.PostalCode = []string{postalCode}
	}
	return name
}

// default template for X509 certificates
func x509Template() x509.Certificate {

	// generate a serial number
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, _ := rand.Int(rand.Reader, serialNumberLimit)

	// set expiry to around 10 years
	expiry := 3650 * 24 * time.Hour
	// round minute and backdate 5 minutes
	notBefore := time.Now().Round(time.Minute).Add(-5 * time.Minute).UTC()

	//basic template to use
	x509 := x509.Certificate{
		SerialNumber:          serialNumber,
		NotBefore:             notBefore,
		NotAfter:              notBefore.Add(expiry).UTC(),
		BasicConstraintsValid: true,
	}
	return x509

}

// generate a signed X509 certificate using ECDSA
func genCertificateECDSA(baseDir, name string, template, parent *x509.Certificate, pub *ecdsa.PublicKey,
	priv interface{}) (*x509.Certificate, error) {

	//create the x509 public cert
	certBytes, err := x509.CreateCertificate(rand.Reader, template, parent, pub, priv)
	if err != nil {
		return nil, err
	}

	//write cert out to file
	fileName := filepath.Join(baseDir, name+"-cert.pem")
	certFile, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}
	//pem encode the cert
	err = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	certFile.Close()
	if err != nil {
		return nil, err
	}

	x509Cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return nil, err
	}
	return x509Cert, nil
}

// LoadCertificateECDSA load a ecdsa cert from a file in cert path
func LoadCertificateECDSA(certPath string) (*x509.Certificate, error) {
	var cert *x509.Certificate
	var err error

	walkFunc := func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".pem") {
			rawCert, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			block, _ := pem.Decode(rawCert)
			cert, err = utils.DERToX509Certificate(block.Bytes)
		}
		return nil
	}

	err = filepath.Walk(certPath, walkFunc)
	if err != nil {
		return nil, err
	}

	return cert, err
}

//generate a signed X509 certficate using GMSM2
func genCertificateGMSM2(baseDir, name string, template, parent *sm2.Certificate, pub *sm2.PublicKey, signerRoot,
	bccspSigner interface{}) (*sm2.Certificate, error) {
	//create the x509 public cert
	certBytes, err := gm.CreateCertificateToMem(template, parent, signerRoot)
	if err != nil {
		return nil, err
	}
	signer := bccspSigner.(*signer.BccspCryptoSigner)

	switch signer.CSP.(type) {
	case *cncc.Impl:
		impl := signer.CSP.(*cncc.Impl)
		err := impl.Uploadcert(template.SubjectKeyId, certBytes)
		if err != nil {
			panic("upload cert error")
		}
	}

	//write cert out to file
	fileName := filepath.Join(baseDir, name+"-cert.pem")
	err = ioutil.WriteFile(fileName, certBytes, os.FileMode(0666))

	// certFile, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}

	// //pem encode the cert
	// err = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	// certFile.Close()
	// if err != nil {
	// 	return nil, err
	// }
	//x509Cert, err := sm2.ReadCertificateFromPem(fileName)

	x509Cert, err := sm2.ReadCertificateFromMem(certBytes)
	if err != nil {
		return nil, err
	}
	// pem格式证书转化成der格式
	block, _ := pem.Decode(certBytes)
	ioutil.WriteFile(filepath.Join(baseDir, name+"-cert.cer"), block.Bytes, os.FileMode(0666))
	return x509Cert, nil

}

//generate a signed X509 certficate using GMSM2
func genRootCertificateGMSM2(baseDir, name string, template, parent *sm2.Certificate, pub *sm2.PublicKey,
	signerRoot interface{}) (*sm2.Certificate, error) {
	//create the x509 public cert
	certBytes, err := gm.CreateCertificateToMem(template, parent, signerRoot)
	if err != nil {
		return nil, err
	}

	//impl := (cncc.Impl)(signer.CSP)
	//
	//err = impl.Uploadcert(template.SubjectKeyId, certBytes)

	if err != nil {
		return nil, err
	}

	//write cert out to file
	fileName := filepath.Join(baseDir, name+"-cert.pem")
	err = ioutil.WriteFile(fileName, certBytes, os.FileMode(0666))

	// certFile, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}

	// //pem encode the cert
	// err = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	// certFile.Close()
	// if err != nil {
	// 	return nil, err
	// }
	//x509Cert, err := sm2.ReadCertificateFromPem(fileName)

	x509Cert, err := sm2.ReadCertificateFromMem(certBytes)
	if err != nil {
		return nil, err
	}
	return x509Cert, nil

}

// LoadCertificateGMSM2 load a ecdsa cert from a file in cert path
func LoadCertificateGMSM2(certPath string) (*sm2.Certificate, error) {
	var cert *sm2.Certificate
	var err error

	walkFunc := func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".pem") {
			rawCert, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			block, _ := pem.Decode(rawCert)
			cert, err = utils.DERToSM2Certificate(block.Bytes)
		}
		return nil
	}

	err = filepath.Walk(certPath, walkFunc)
	if err != nil {
		return nil, err
	}

	return cert, err
}