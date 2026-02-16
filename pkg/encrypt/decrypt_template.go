package encrypt

import (
	"encoding/base64"
	"fmt"
	"strings"
)

// DefaultDecryptPageHTML is the default decrypt.html template created at the site root.
// Structured like other site-level templates (header.html, footer.html) with full HTML structure.
// Only the <body> content is used when building the encrypted page.
const DefaultDecryptPageHTML = `<!doctype html>
<html>
<head></head>
<body>
<div style="display:flex;justify-content:center;align-items:center;min-height:100vh;font-family:sans-serif">
  <form id="decrypt-form" style="text-align:center">
    <h2>This page is encrypted</h2>
    <p>Enter the passphrase to view this page.</p>
    <input id="decrypt-password" type="password" placeholder="Passphrase" autofocus
      style="padding:8px 12px;font-size:16px;border:1px solid #ccc;border-radius:4px;margin-right:8px">
    <button type="submit" style="padding:8px 16px;font-size:16px;cursor:pointer;border:1px solid #ccc;border-radius:4px;background:#f5f5f5">Decrypt</button>
    <p id="decrypt-error" style="color:red;display:none">Wrong passphrase. Please try again.</p>
  </form>
</div>
</body>
</html>`

// BuildEncryptedPage assembles a complete encrypted HTML page.
// The decryptFormHTML is the content of components/decrypt.html (or DefaultDecryptFormHTML).
func BuildEncryptedPage(salt, iv, ciphertext []byte, decryptFormHTML string) string {
	saltB64 := base64.StdEncoding.EncodeToString(salt)
	ivB64 := base64.StdEncoding.EncodeToString(iv)
	ctB64 := base64.StdEncoding.EncodeToString(ciphertext)

	return fmt.Sprintf(`<!doctype html>
<html>
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Encrypted Page</title>
</head>
<body>
%s
<script id="encrypted-payload" type="application/json">%s</script>
<script>
(function() {
  var SALT = "%s";
  var IV = "%s";
  var ITERATIONS = 100000;

  function b64ToBytes(b64) {
    var bin = atob(b64);
    var bytes = new Uint8Array(bin.length);
    for (var i = 0; i < bin.length; i++) bytes[i] = bin.charCodeAt(i);
    return bytes;
  }

  async function decrypt(passphrase) {
    var salt = b64ToBytes(SALT);
    var iv = b64ToBytes(IV);
    var ct = b64ToBytes(document.getElementById("encrypted-payload").textContent);

    var enc = new TextEncoder();
    var keyMaterial = await crypto.subtle.importKey("raw", enc.encode(passphrase), "PBKDF2", false, ["deriveKey"]);
    var key = await crypto.subtle.deriveKey(
      {name: "PBKDF2", salt: salt, iterations: ITERATIONS, hash: "SHA-256"},
      keyMaterial,
      {name: "AES-GCM", length: 256},
      false,
      ["decrypt"]
    );

    var decrypted = await crypto.subtle.decrypt({name: "AES-GCM", iv: iv}, key, ct);
    return new TextDecoder().decode(decrypted);
  }

  var form = document.getElementById("decrypt-form");
  if (form) {
    form.addEventListener("submit", async function(e) {
      e.preventDefault();
      var pw = document.getElementById("decrypt-password").value;
      try {
        var html = await decrypt(pw);
        document.open();
        document.write(html);
        document.close();
      } catch(err) {
        var el = document.getElementById("decrypt-error");
        if (el) { el.style.display = "block"; }
      }
    });
  }
})();
</script>
</body>
</html>`, strings.TrimSpace(decryptFormHTML), ctB64, saltB64, ivB64)
}
