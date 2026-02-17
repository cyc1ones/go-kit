package x509util

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRsa(t *testing.T) {
	// 生成密钥对
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	pubKey := &privKey.PublicKey

	plainText := []byte("hello world")

	// 用初始公钥加密一段数据
	cipher, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, plainText)
	require.NoError(t, err)

	var privKey2 *rsa.PrivateKey
	var pubKey2 *rsa.PublicKey

	{
		// 编码私钥
		pemPrivKey, err := EncodePKCS8PrivateKey(privKey)
		require.NoError(t, err)

		// 编码公钥
		pemPubKey, err := EncodePKIXPublicKey(pubKey)
		require.NoError(t, err)

		var ok bool
		// 解码私钥
		decodedPrivKeyRaw, err := DecodePKCS8PrivateKey(pemPrivKey)
		require.NoError(t, err)
		privKey2, ok = decodedPrivKeyRaw.(*rsa.PrivateKey)
		require.True(t, ok, "decoded private key is not *rsa.PrivateKey")

		// 解码公钥
		decodedPubKeyRaw, err := DecodePKIXPublicKey(pemPubKey)
		require.NoError(t, err)
		pubKey2, ok = decodedPubKeyRaw.(*rsa.PublicKey)
		require.True(t, ok, "decoded public key is not *rsa.PublicKey")
	}

	// 用编码后解码的私钥解密刚刚加密的数据
	decrypted, err := rsa.DecryptPKCS1v15(rand.Reader, privKey2, cipher)
	require.NoError(t, err)
	require.Equal(t, plainText, decrypted)

	// 用编码后解码的公钥加密，原始私钥解密
	cipher, err = rsa.EncryptPKCS1v15(rand.Reader, pubKey2, plainText)
	require.NoError(t, err)
	decrypted, err = rsa.DecryptPKCS1v15(rand.Reader, privKey, cipher)
	require.NoError(t, err)
	require.Equal(t, plainText, decrypted)
}
