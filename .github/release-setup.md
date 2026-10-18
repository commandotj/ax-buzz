# macOS release setup

Buzz release workflow publishes notarized macOS DMGs only. CLI ships inside `Buzz.app`; no standalone CLI archive is released. The `macos-release` environment accepts `v*` tags. Apple credentials belong in that environment, never in repository files.

## Before you start

You need an active Apple Developer Program team and access to its Apple Developer account and App Store Connect. Creating a Developer ID certificate requires the team's Account Holder. Buzz ships a DMG, so create a **Developer ID Application** certificate. Do not create a Developer ID Installer certificate; Buzz does not ship a PKG.

## Get signing certificate

1. Open [Certificates, Identifiers & Profiles](https://developer.apple.com/account/resources/certificates/list). Choose **Certificates** → **+** → **Developer ID** → **Developer ID Application**.
2. Create a certificate signing request on your Mac in Keychain Access. Upload the `.certSigningRequest`, download Apple's `.cer`, and open it to install it beside its private key.
3. In Keychain Access → **My Certificates**, confirm the Developer ID certificate expands to show its private key. Export that identity as a password-protected `.p12` file. The `.cer` alone cannot sign builds.
4. Find the exact identity and Team ID with `security find-identity -v -p codesigning`. Record the complete `Developer ID Application: ... (TEAMID)` identity. Confirm the Team ID in Apple Developer account **Membership details**.

## Get notarization login

1. In [App Store Connect](https://appstoreconnect.apple.com/), open **Users and Access** → **Integrations** → **App Store Connect API** → **Team Keys**. Create a team key for Buzz notarization. Individual keys cannot authenticate `notarytool`.
2. Save the downloaded `.p8` key securely. Apple does not let you download that private key again. Record its Key ID and Issuer ID from the same page. Individual API keys cannot authenticate `notarytool`; create a Team Key. Team Keys apply across the developer team, so ask the team administrator to confirm the key's role permits notarization and is appropriate for that scope.

## Add GitHub secrets

Open repository **Settings** → **Environments** → **macos-release** → **Environment secrets**. Add these exact names:

| Secret | Value |
| --- | --- |
| `APPLE_CERTIFICATE_P12_BASE64` | One-line base64 encoding of password-protected Developer ID `.p12` |
| `APPLE_CERTIFICATE_PASSWORD` | Password used when exporting the `.p12` |
| `APPLE_SIGNING_IDENTITY` | Exact `Developer ID Application: ... (TEAMID)` identity |
| `APPLE_TEAM_ID` | Team ID shown in Apple Developer membership details |
| `APPLE_NOTARY_API_KEY_P8_BASE64` | One-line base64 encoding of downloaded team `.p8` key |
| `APPLE_NOTARY_API_KEY_ID` | Key ID for the team API key |
| `APPLE_NOTARY_ISSUER_ID` | Issuer ID for the team API key |

On your Mac, encode files locally and paste the output directly into the matching GitHub secret form:

```sh
base64 -i DeveloperID.p12 | tr -d '\n'
base64 -i AuthKey_KEYID.p8 | tr -d '\n'
```

Never commit these files or send encoded values in chat. The workflow checks for missing values, imports the certificate into a temporary runner keychain, checks its identity and Team ID, and validates the notary credentials before building.

## Run and check release

After the version changes are merged into `main`, create and push a semantic version tag such as `v1.2.3`. Workflow checks that the tag points to `main`, verifies app and bundled CLI versions, builds native arm64 and Intel DMGs, signs and notarizes both, staples and validates them, then publishes only the two DMGs. The workflow fails before publication if any credential or validation is missing.

Inspect the **Buzz.app release** workflow run. Both `Build arm64 DMG` and `Build amd64 DMG` must pass signing, notarization, Gatekeeper, mounted-image, version, and architecture checks. `Publish GitHub Release` runs only after both pass. SHA-256 values appear in release notes.

## References

- [Apple Developer ID certificates](https://developer.apple.com/help/account/certificates/create-developer-id-certificates/)
- [Apple App Store Connect API keys](https://developer.apple.com/help/app-store-connect/get-started/app-store-connect-api/)
- [Apple notarytool workflow](https://developer.apple.com/documentation/security/customizing-the-notarization-workflow)
- [Wails macOS packaging](https://v3.wails.io/guides/build/macos/)
