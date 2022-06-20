package web

import (
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/smartcontractkit/chainlink/core/logger/audit"
	"github.com/smartcontractkit/chainlink/core/services/chainlink"
	"github.com/smartcontractkit/chainlink/core/services/keystore/keys/p2pkey"
	"github.com/smartcontractkit/chainlink/core/web/presenters"
)

// P2PKeysController manages P2P keys
type P2PKeysController struct {
	App chainlink.Application
}

// Index lists P2P keys
// Example:
// "GET <application>/keys/p2p"
func (p2pkc *P2PKeysController) Index(c *gin.Context) {
	keys, err := p2pkc.App.GetKeyStore().P2P().GetAll()
	if err != nil {
		jsonAPIError(c, http.StatusInternalServerError, err)
		return
	}
	jsonAPIResponse(c, presenters.NewP2PKeyResources(keys), "p2pKey")
}

// Create and return a P2P key
// Example:
// "POST <application>/keys/p2p"
func (p2pkc *P2PKeysController) Create(c *gin.Context) {
	key, err := p2pkc.App.GetKeyStore().P2P().Create()
	if err != nil {
		jsonAPIError(c, http.StatusInternalServerError, err)
		return
	}

	p2pkc.App.GetLogger().Audit(audit.P2PKeyCreated, map[string]interface{}{
		"p2pPublicKey": key.PublicKeyHex(),
		"p2pID":        key.ID(),
		"p2pPeerID":    key.PeerID(),
		"p2pType":      key.Type(),
	})
	jsonAPIResponse(c, presenters.NewP2PKeyResource(key), "p2pKey")
}

// Delete a P2P key
// Example:
// "DELETE <application>/keys/p2p/:keyID"
// "DELETE <application>/keys/p2p/:keyID?hard=true"
func (p2pkc *P2PKeysController) Delete(c *gin.Context) {
	keyID, err := p2pkey.MakePeerID(c.Param("keyID"))
	if err != nil {
		jsonAPIError(c, http.StatusUnprocessableEntity, err)
		return
	}
	key, err := p2pkc.App.GetKeyStore().P2P().Get(keyID)
	if err != nil {
		jsonAPIError(c, http.StatusNotFound, err)
		return
	}
	_, err = p2pkc.App.GetKeyStore().P2P().Delete(key.PeerID())
	if err != nil {
		jsonAPIError(c, http.StatusInternalServerError, err)
		return
	}

	p2pkc.App.GetLogger().Audit(audit.P2PKeyDeleted, map[string]interface{}{"id": keyID})
	jsonAPIResponse(c, presenters.NewP2PKeyResource(key), "p2pKey")
}

// Import imports a P2P key
// Example:
// "Post <application>/keys/p2p/import"
func (p2pkc *P2PKeysController) Import(c *gin.Context) {
	defer p2pkc.App.GetLogger().ErrorIfClosing(c.Request.Body, "Import ")

	bytes, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		jsonAPIError(c, http.StatusBadRequest, err)
		return
	}
	oldPassword := c.Query("oldpassword")
	key, err := p2pkc.App.GetKeyStore().P2P().Import(bytes, oldPassword)
	if err != nil {
		jsonAPIError(c, http.StatusInternalServerError, err)
		return
	}

	p2pkc.App.GetLogger().Audit(audit.P2PKeyImported, map[string]interface{}{
		"p2pPublicKey": key.PublicKeyHex(),
		"p2pID":        key.ID(),
		"p2pPeerID":    key.PeerID(),
		"p2pType":      key.Type(),
	})
	jsonAPIResponse(c, presenters.NewP2PKeyResource(key), "p2pKey")
}

// Export exports a P2P key
// Example:
// "Post <application>/keys/p2p/export"
func (p2pkc *P2PKeysController) Export(c *gin.Context) {
	defer p2pkc.App.GetLogger().ErrorIfClosing(c.Request.Body, "Export request body")

	keyID, err := p2pkey.MakePeerID(c.Param("ID"))
	if err != nil {
		jsonAPIError(c, http.StatusUnprocessableEntity, err)
		return
	}

	newPassword := c.Query("newpassword")
	bytes, err := p2pkc.App.GetKeyStore().P2P().Export(keyID, newPassword)
	if err != nil {
		jsonAPIError(c, http.StatusInternalServerError, err)
		return
	}

	p2pkc.App.GetLogger().Audit(audit.P2PKeyExported, map[string]interface{}{"keyID": keyID})
	c.Data(http.StatusOK, MediaType, bytes)
}
