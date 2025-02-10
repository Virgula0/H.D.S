$(function () {
    // ------------------------------------------------
    // LOADING OVERLAY
    // ------------------------------------------------
    const $loadingOverlay = $("#loadingOverlay");
    setTimeout(() => $loadingOverlay.fadeOut(400), 500);

    const showLoading = () => $loadingOverlay.show();
    const hideLoading = () => $loadingOverlay.fadeOut(300);

    // ------------------------------------------------
    // DARK MODE
    // ------------------------------------------------
    const setDarkMode = (isDark) => {
        document.documentElement.classList.toggle("dark-mode", isDark);
        document.body.classList.toggle("dark-mode", isDark);
        localStorage.setItem("darkMode", isDark);

        const $themeToggle = $("#theme-toggle i");
        if (isDark) {
            $themeToggle.removeClass("fa-moon").addClass("fa-sun");
        } else {
            $themeToggle.removeClass("fa-sun").addClass("fa-moon");
        }
    };

    const isDarkMode = () => localStorage.getItem("darkMode") === "true";

    // Apply dark mode without transition to avoid flicker on page load
    const applyDarkModeWithoutTransition = (isDark) => {
        document.documentElement.style.transition = "none";
        document.body.style.transition = "none";
        setDarkMode(isDark);
        // Force reflow to apply changes immediately
        document.documentElement.offsetHeight;
        document.documentElement.style.transition = "";
        document.body.style.transition = "";
    };

    applyDarkModeWithoutTransition(isDarkMode());

    // Fade in main content
    $("#page-content-wrapper").addClass("loaded");

    // ------------------------------------------------
    // SIDEBAR & THEME TOGGLES
    // ------------------------------------------------
    $(document).on("click", "#menu-toggle", (e) => {
        e.preventDefault();
        $("#wrapper").toggleClass("toggled");
        $("body").toggleClass("sidebar-open");
    });

    $(document).on("click", "#theme-toggle", () => {
        setDarkMode(!isDarkMode());
    });

    // ------------------------------------------------
    // SELECT POPULATION (HANDSHAKE-RELATED)
    // ------------------------------------------------
    const populateSelect = (selector, options, placeholder = "") => {
        const $select = $(selector);
        if ($select.length) {
            $select.empty();
            if (placeholder) {
                $select.append(`<option value="">${placeholder}</option>`);
            }
            options.forEach((opt) =>
                $select.append(`<option value="${opt.value}">${opt.label}</option>`)
            );
        }
    };

    const attackModes = [
        { value: 0, label: "Straight" },
        { value: 1, label: "Combination" },
        { value: 3, label: "Brute-force" },
        { value: 6, label: "Hybrid Wordlist + Mask" },
        { value: 7, label: "Hybrid Mask + Wordlist" },
    ];

    const hashModes = [
        { value: 900,  label: "MD4 | Raw Hash" },
        { value: 0,    label: "MD5 | Raw Hash" },
        { value: 100,  label: "SHA1 | Raw Hash" },
        { value: 1300, label: "SHA2-224 | Raw Hash" },
        { value: 1400, label: "SHA2-256 | Raw Hash" },
        { value: 10800, label: "SHA2-384 | Raw Hash" },
        { value: 1700, label: "SHA2-512 | Raw Hash" },
        { value: 17300, label: "SHA3-224 | Raw Hash" },
        { value: 17400, label: "SHA3-256 | Raw Hash" },
        { value: 17500, label: "SHA3-384 | Raw Hash" },
        { value: 17600, label: "SHA3-512 | Raw Hash" },
        { value: 6000,  label: "RIPEMD-160 | Raw Hash" },
        { value: 600,   label: "BLAKE2b-512 | Raw Hash" },
        { value: 11700, label: "GOST R 34.11-2012 (Streebog) 256-bit, big-endian | Raw Hash" },
        { value: 11800, label: "GOST R 34.11-2012 (Streebog) 512-bit, big-endian | Raw Hash" },
        { value: 6900,  label: "GOST R 34.11-94 | Raw Hash" },
        { value: 5100,  label: "Half MD5 | Raw Hash" },
        { value: 18700, label: "Java Object hashCode() | Raw Hash" },
        { value: 17700, label: "Keccak-224 | Raw Hash" },
        { value: 17800, label: "Keccak-256 | Raw Hash" },
        { value: 17900, label: "Keccak-384 | Raw Hash" },
        { value: 18000, label: "Keccak-512 | Raw Hash" },
        { value: 21400, label: "sha256(sha256_bin($pass)) | Raw Hash" },
        { value: 6100,  label: "Whirlpool | Raw Hash" },
        { value: 10100, label: "SipHash | Raw Hash" },
        { value: 21000, label: "BitShares v0.x - sha512(sha512_bin(pass)) | Raw Hash" },

        // Raw Hash, Salted and/or Iterated
        { value: 10,   label: "md5($pass.$salt) | Raw Hash, Salted/Iterated" },
        { value: 20,   label: "md5($salt.$pass) | Raw Hash, Salted/Iterated" },
        { value: 3800, label: "md5($salt.$pass.$salt) | Raw Hash, Salted/Iterated" },
        { value: 3710, label: "md5($salt.md5($pass)) | Raw Hash, Salted/Iterated" },
        { value: 4110, label: "md5($salt.md5($pass.$salt)) | Raw Hash, Salted/Iterated" },
        { value: 4010, label: "md5($salt.md5($salt.$pass)) | Raw Hash, Salted/Iterated" },
        { value: 21300, label: "md5($salt.sha1($salt.$pass)) | Raw Hash, Salted/Iterated" },
        { value: 40,   label: "md5($salt.utf16le($pass)) | Raw Hash, Salted/Iterated" },
        { value: 2600, label: "md5(md5($pass)) | Raw Hash, Salted/Iterated" },
        { value: 3910, label: "md5(md5($pass).md5($salt)) | Raw Hash, Salted/Iterated" },
        { value: 4400, label: "md5(sha1($pass)) | Raw Hash, Salted/Iterated" },
        { value: 20900, label: "md5(sha1($pass).md5($pass).sha1($pass)) | Raw Hash, Salted/Iterated" },
        { value: 21200, label: "md5(sha1($salt).md5($pass)) | Raw Hash, Salted/Iterated" },
        { value: 4300, label: "md5(strtoupper(md5($pass))) | Raw Hash, Salted/Iterated" },
        { value: 30,   label: "md5(utf16le($pass).$salt) | Raw Hash, Salted/Iterated" },

        { value: 110,  label: "sha1($pass.$salt) | Raw Hash, Salted/Iterated" },
        { value: 120,  label: "sha1($salt.$pass) | Raw Hash, Salted/Iterated" },
        { value: 4900, label: "sha1($salt.$pass.$salt) | Raw Hash, Salted/Iterated" },
        { value: 4520, label: "sha1($salt.sha1($pass)) | Raw Hash, Salted/Iterated" },
        { value: 140,  label: "sha1($salt.utf16le($pass)) | Raw Hash, Salted/Iterated" },
        { value: 19300, label: "sha1($salt1.$pass.$salt2) | Raw Hash, Salted/Iterated" },
        { value: 14400, label: "sha1(CX) | Raw Hash, Salted/Iterated" },
        { value: 4700, label: "sha1(md5($pass)) | Raw Hash, Salted/Iterated" },
        { value: 4710, label: "sha1(md5($pass).$salt) | Raw Hash, Salted/Iterated" },
        { value: 21100, label: "sha1(md5($pass.$salt)) | Raw Hash, Salted/Iterated" },
        { value: 18500, label: "sha1(md5(md5($pass))) | Raw Hash, Salted/Iterated" },
        { value: 4500,  label: "sha1(sha1($pass)) | Raw Hash, Salted/Iterated" },
        { value: 130,   label: "sha1(utf16le($pass).$salt) | Raw Hash, Salted/Iterated" },

        { value: 1410, label: "sha256($pass.$salt) | Raw Hash, Salted/Iterated" },
        { value: 1420, label: "sha256($salt.$pass) | Raw Hash, Salted/Iterated" },
        { value: 22300, label: "sha256($salt.$pass.$salt) | Raw Hash, Salted/Iterated" },
        { value: 1440, label: "sha256($salt.utf16le($pass)) | Raw Hash, Salted/Iterated" },
        { value: 20800, label: "sha256(md5($pass)) | Raw Hash, Salted/Iterated" },
        { value: 20710, label: "sha256(sha256($pass).$salt) | Raw Hash, Salted/Iterated" },
        { value: 1430, label: "sha256(utf16le($pass).$salt) | Raw Hash, Salted/Iterated" },

        { value: 1710, label: "sha512($pass.$salt) | Raw Hash, Salted/Iterated" },
        { value: 1720, label: "sha512($salt.$pass) | Raw Hash, Salted/Iterated" },
        { value: 1740, label: "sha512($salt.utf16le($pass)) | Raw Hash, Salted/Iterated" },
        { value: 1730, label: "sha512(utf16le($pass).$salt) | Raw Hash, Salted/Iterated" },
        { value: 19500, label: "Ruby on Rails Restful-Authentication | Raw Hash, Salted/Iterated" },

        // HMAC (key = pass or salt)
        { value: 50,   label: "HMAC-MD5 (key = $pass) | Raw Hash, Authenticated" },
        { value: 60,   label: "HMAC-MD5 (key = $salt) | Raw Hash, Authenticated" },
        { value: 150,  label: "HMAC-SHA1 (key = $pass) | Raw Hash, Authenticated" },
        { value: 160,  label: "HMAC-SHA1 (key = $salt) | Raw Hash, Authenticated" },
        { value: 1450, label: "HMAC-SHA256 (key = $pass) | Raw Hash, Authenticated" },
        { value: 1460, label: "HMAC-SHA256 (key = $salt) | Raw Hash, Authenticated" },
        { value: 1750, label: "HMAC-SHA512 (key = $pass) | Raw Hash, Authenticated" },
        { value: 1760, label: "HMAC-SHA512 (key = $salt) | Raw Hash, Authenticated" },
        { value: 11750, label: "HMAC-Streebog-256 (key = $pass), big-endian | Raw Hash, Authenticated" },
        { value: 11760, label: "HMAC-Streebog-256 (key = $salt), big-endian | Raw Hash, Authenticated" },
        { value: 11850, label: "HMAC-Streebog-512 (key = $pass), big-endian | Raw Hash, Authenticated" },
        { value: 11860, label: "HMAC-Streebog-512 (key = $salt), big-endian | Raw Hash, Authenticated" },

        // Raw Checksum
        { value: 11500, label: "CRC32 | Raw Checksum" },

        // Raw Cipher, Known-Plaintext attack
        { value: 14100, label: "3DES (PT = $salt, key = $pass) | Raw Cipher, Known-Plaintext" },
        { value: 14000, label: "DES (PT = $salt, key = $pass) | Raw Cipher, Known-Plaintext" },
        { value: 15400, label: "ChaCha20 | Raw Cipher, Known-Plaintext" },
        { value: 14900, label: "Skip32 (PT = $salt, key = $pass) | Raw Cipher, Known-Plaintext" },

        // Generic KDF
        { value: 11900, label: "PBKDF2-HMAC-MD5 | Generic KDF" },
        { value: 12000, label: "PBKDF2-HMAC-SHA1 | Generic KDF" },
        { value: 10900, label: "PBKDF2-HMAC-SHA256 | Generic KDF" },
        { value: 12100, label: "PBKDF2-HMAC-SHA512 | Generic KDF" },
        { value: 8900,  label: "scrypt | Generic KDF" },
        { value: 400,   label: "phpass | Generic KDF" },
        { value: 16900, label: "Ansible Vault | Generic KDF" },
        { value: 12001, label: "Atlassian (PBKDF2-HMAC-SHA1) | Generic KDF" },
        { value: 20200, label: "Python passlib pbkdf2-sha512 | Generic KDF" },
        { value: 20300, label: "Python passlib pbkdf2-sha256 | Generic KDF" },
        { value: 20400, label: "Python passlib pbkdf2-sha1 | Generic KDF" },

        // Network Protocols
        { value: 16100, label: "TACACS+ | Network Protocols" },
        { value: 11400, label: "SIP digest authentication (MD5) | Network Protocols" },
        { value: 5300,  label: "IKE-PSK MD5 | Network Protocols" },
        { value: 5400,  label: "IKE-PSK SHA1 | Network Protocols" },
        { value: 23200, label: "XMPP SCRAM PBKDF2-SHA1 | Network Protocols" },
        { value: 2500,  label: "WPA-EAPOL-PBKDF2 | Network Protocols" },
        { value: 2501,  label: "WPA-EAPOL-PMK | Network Protocols" },
        { value: 22000, label: "WPA-PBKDF2-PMKID+EAPOL | Network Protocols" },
        { value: 22001, label: "WPA-PMK-PMKID+EAPOL | Network Protocols" },
        { value: 16800, label: "WPA-PMKID-PBKDF2 | Network Protocols" },
        { value: 16801, label: "WPA-PMKID-PMK | Network Protocols" },
        { value: 7300,  label: "IPMI2 RAKP HMAC-SHA1 | Network Protocols" },
        { value: 10200, label: "CRAM-MD5 | Network Protocols" },
        { value: 4800,  label: "iSCSI CHAP authentication, MD5(CHAP) | Network Protocols" },
        { value: 16500, label: "JWT (JSON Web Token) | Network Protocols" },
        { value: 22600, label: "Telegram Desktop App Passcode (PBKDF2-HMAC-SHA1) | Network Protocols" },
        { value: 22301, label: "Telegram Mobile App Passcode (SHA256) | Network Protocols" },
        { value: 7500,  label: "Kerberos 5, etype 23, AS-REQ Pre-Auth | Network Protocols" },
        { value: 13100, label: "Kerberos 5, etype 23, TGS-REP | Network Protocols" },
        { value: 18200, label: "Kerberos 5, etype 23, AS-REP | Network Protocols" },
        { value: 19600, label: "Kerberos 5, etype 17, TGS-REP | Network Protocols" },
        { value: 19700, label: "Kerberos 5, etype 18, TGS-REP | Network Protocols" },
        { value: 19800, label: "Kerberos 5, etype 17, Pre-Auth | Network Protocols" },
        { value: 19900, label: "Kerberos 5, etype 18, Pre-Auth | Network Protocols" },
        { value: 5500,  label: "NetNTLMv1 / NetNTLMv1+ESS | Network Protocols" },
        { value: 5600,  label: "NetNTLMv2 | Network Protocols" },
        { value: 23,    label: "Skype | Network Protocols" },
        { value: 11100, label: "PostgreSQL CRAM (MD5) | Network Protocols" },
        { value: 11200, label: "MySQL CRAM (SHA1) | Network Protocols" },

        // Operating System
        { value: 8500,  label: "RACF | Operating System" },
        { value: 6300,  label: "AIX {smd5} | Operating System" },
        { value: 6700,  label: "AIX {ssha1} | Operating System" },
        { value: 6400,  label: "AIX {ssha256} | Operating System" },
        { value: 6500,  label: "AIX {ssha512} | Operating System" },
        { value: 3000,  label: "LM | Operating System" },
        { value: 19000, label: "QNX /etc/shadow (MD5) | Operating System" },
        { value: 19100, label: "QNX /etc/shadow (SHA256) | Operating System" },
        { value: 19200, label: "QNX /etc/shadow (SHA512) | Operating System" },
        { value: 15300, label: "DPAPI masterkey file v1 | Operating System" },
        { value: 15900, label: "DPAPI masterkey file v2 | Operating System" },
        { value: 7200,  label: "GRUB 2 | Operating System" },
        { value: 12800, label: "MS-AzureSync PBKDF2-HMAC-SHA256 | Operating System" },
        { value: 12400, label: "BSDi Crypt, Extended DES | Operating System" },
        { value: 1000,  label: "NTLM | Operating System" },
        { value: 122,   label: "macOS v10.4, macOS v10.5, MacOS v10.6 | Operating System" },
        { value: 1722,  label: "macOS v10.7 | Operating System" },
        { value: 7100,  label: "macOS v10.8+ (PBKDF2-SHA512) | Operating System" },
        { value: 9900,  label: "Radmin2 | Operating System" },
        { value: 5800,  label: "Samsung Android Password/PIN | Operating System" },
        { value: 3200,  label: "bcrypt $2*$, Blowfish (Unix) | Operating System" },
        { value: 500,   label: "md5crypt, MD5 (Unix), Cisco-IOS $1$ (MD5) | Operating System" },
        { value: 1500,  label: "descrypt, DES (Unix), Traditional DES | Operating System" },
        { value: 7400,  label: "sha256crypt $5$, SHA256 (Unix) | Operating System" },
        { value: 1800,  label: "sha512crypt $6$, SHA512 (Unix) | Operating System" },
        { value: 13800, label: "Windows Phone 8+ PIN/password | Operating System" },
        { value: 2410,  label: "Cisco-ASA MD5 | Operating System" },
        { value: 9200,  label: "Cisco-IOS $8$ (PBKDF2-SHA256) | Operating System" },
        { value: 9300,  label: "Cisco-IOS $9$ (scrypt) | Operating System" },
        { value: 5700,  label: "Cisco-IOS type 4 (SHA256) | Operating System" },
        { value: 2400,  label: "Cisco-PIX MD5 | Operating System" },
        { value: 8100,  label: "Citrix NetScaler (SHA1) | Operating System" },
        { value: 22200, label: "Citrix NetScaler (SHA512) | Operating System" },
        { value: 1100,  label: "Domain Cached Credentials (DCC), MS Cache | Operating System" },
        { value: 2100,  label: "Domain Cached Credentials 2 (DCC2), MS Cache 2 | Operating System" },
        { value: 7000,  label: "FortiGate (FortiOS) | Operating System" },
        { value: 125,   label: "ArubaOS | Operating System" },
        { value: 501,   label: "Juniper IVE | Operating System" },
        { value: 22,    label: "Juniper NetScreen/SSG (ScreenOS) | Operating System" },
        { value: 15100, label: "Juniper/NetBSD sha1crypt | Operating System" },

        // Database Server
        { value: 131,   label: "MSSQL (2000) | Database Server" },
        { value: 132,   label: "MSSQL (2005) | Database Server" },
        { value: 1731,  label: "MSSQL (2012, 2014) | Database Server" },
        { value: 12,    label: "PostgreSQL | Database Server" },
        { value: 3100,  label: "Oracle H: Type (Oracle 7+) | Database Server" },
        { value: 112,   label: "Oracle S: Type (Oracle 11+) | Database Server" },
        { value: 12300, label: "Oracle T: Type (Oracle 12+) | Database Server" },
        { value: 7401,  label: "MySQL $A$ (sha256crypt) | Database Server" },
        { value: 200,   label: "MySQL323 | Database Server" },
        { value: 300,   label: "MySQL4.1/MySQL5 | Database Server" },
        { value: 8000,  label: "Sybase ASE | Database Server" },
        { value: 1421,  label: "hMailServer | FTP, HTTP, SMTP, LDAP Server" },
        { value: 8300,  label: "DNSSEC (NSEC3) | FTP, HTTP, SMTP, LDAP Server" },
        { value: 16400, label: "CRAM-MD5 Dovecot | FTP, HTTP, SMTP, LDAP Server" },
        { value: 1411,  label: "SSHA-256(Base64), LDAP {SSHA256} | FTP, HTTP, SMTP, LDAP Server" },
        { value: 1711,  label: "SSHA-512(Base64), LDAP {SSHA512} | FTP, HTTP, SMTP, LDAP Server" },
        { value: 10901, label: "RedHat 389-DS LDAP (PBKDF2-HMAC-SHA256) | FTP, HTTP, SMTP, LDAP Server" },
        { value: 15000, label: "FileZilla Server >= 0.9.55 | FTP, HTTP, SMTP, LDAP Server" },
        { value: 12600, label: "ColdFusion 10+ | FTP, HTTP, SMTP, LDAP Server" },
        { value: 1600,  label: "Apache $apr1$ MD5, md5apr1, MD5 (APR) | FTP, HTTP, SMTP, LDAP Server" },
        { value: 141,   label: "Episerver 6.x < .NET 4 | FTP, HTTP, SMTP, LDAP Server" },
        { value: 1441,  label: "Episerver 6.x >= .NET 4 | FTP, HTTP, SMTP, LDAP Server" },
        { value: 101,   label: "nsldap, SHA-1(Base64), Netscape LDAP SHA | FTP, HTTP, SMTP, LDAP Server" },
        { value: 111,   label: "nsldaps, SSHA-1(Base64), Netscape LDAP SSHA | FTP, HTTP, SMTP, LDAP Server" },

        // Enterprise Application Software (EAS)
        { value: 7700,  label: "SAP CODVN B (BCODE) | EAS" },
        { value: 7701,  label: "SAP CODVN B (BCODE) from RFC_READ_TABLE | EAS" },
        { value: 7800,  label: "SAP CODVN F/G (PASSCODE) | EAS" },
        { value: 7801,  label: "SAP CODVN F/G (PASSCODE) from RFC_READ_TABLE | EAS" },
        { value: 10300, label: "SAP CODVN H (PWDSALTEDHASH) iSSHA-1 | EAS" },
        { value: 133,   label: "PeopleSoft | EAS" },
        { value: 13500, label: "PeopleSoft PS_TOKEN | EAS" },
        { value: 21500, label: "SolarWinds Orion | EAS" },
        { value: 8600,  label: "Lotus Notes/Domino 5 | EAS" },
        { value: 8700,  label: "Lotus Notes/Domino 6 | EAS" },
        { value: 9100,  label: "Lotus Notes/Domino 8 | EAS" },
        { value: 20600, label: "Oracle Transportation Management (SHA256) | EAS" },
        { value: 4711,  label: "Huawei sha1(md5($pass).$salt) | EAS" },
        { value: 20711, label: "AuthMe sha256 | EAS" },

        // Full-Disk Encryption (FDE)
        { value: 12200, label: "eCryptfs | FDE" },
        { value: 22400, label: "AES Crypt (SHA256) | FDE" },
        { value: 14600, label: "LUKS | FDE" },
        { value: 13711, label: "VeraCrypt RIPEMD160 + XTS 512 bit | FDE" },
        { value: 13712, label: "VeraCrypt RIPEMD160 + XTS 1024 bit | FDE" },
        { value: 13713, label: "VeraCrypt RIPEMD160 + XTS 1536 bit | FDE" },
        { value: 13741, label: "VeraCrypt RIPEMD160 + XTS 512 bit + boot-mode | FDE" },
        { value: 13742, label: "VeraCrypt RIPEMD160 + XTS 1024 bit + boot-mode | FDE" },
        { value: 13743, label: "VeraCrypt RIPEMD160 + XTS 1536 bit + boot-mode | FDE" },
        { value: 13751, label: "VeraCrypt SHA256 + XTS 512 bit | FDE" },
        { value: 13752, label: "VeraCrypt SHA256 + XTS 1024 bit | FDE" },
        { value: 13753, label: "VeraCrypt SHA256 + XTS 1536 bit | FDE" },
        { value: 13761, label: "VeraCrypt SHA256 + XTS 512 bit + boot-mode | FDE" },
        { value: 13762, label: "VeraCrypt SHA256 + XTS 1024 bit + boot-mode | FDE" },
        { value: 13763, label: "VeraCrypt SHA256 + XTS 1536 bit + boot-mode | FDE" },
        { value: 13721, label: "VeraCrypt SHA512 + XTS 512 bit | FDE" },
        { value: 13722, label: "VeraCrypt SHA512 + XTS 1024 bit | FDE" },
        { value: 13723, label: "VeraCrypt SHA512 + XTS 1536 bit | FDE" },
        { value: 13771, label: "VeraCrypt Streebog-512 + XTS 512 bit | FDE" },
        { value: 13772, label: "VeraCrypt Streebog-512 + XTS 1024 bit | FDE" },
        { value: 13773, label: "VeraCrypt Streebog-512 + XTS 1536 bit | FDE" },
        { value: 13731, label: "VeraCrypt Whirlpool + XTS 512 bit | FDE" },
        { value: 13732, label: "VeraCrypt Whirlpool + XTS 1024 bit | FDE" },
        { value: 13733, label: "VeraCrypt Whirlpool + XTS 1536 bit | FDE" },
        { value: 16700, label: "FileVault 2 | FDE" },
        { value: 20011, label: "DiskCryptor SHA512 + XTS 512 bit | FDE" },
        { value: 20012, label: "DiskCryptor SHA512 + XTS 1024 bit | FDE" },
        { value: 20013, label: "DiskCryptor SHA512 + XTS 1536 bit | FDE" },
        { value: 22100, label: "BitLocker | FDE" },
        { value: 12900, label: "Android FDE (Samsung DEK) | FDE" },
        { value: 8800,  label: "Android FDE <= 4.3 | FDE" },
        { value: 18300, label: "Apple File System (APFS) | FDE" },
        { value: 6211,  label: "TrueCrypt RIPEMD160 + XTS 512 bit | FDE" },
        { value: 6212,  label: "TrueCrypt RIPEMD160 + XTS 1024 bit | FDE" },
        { value: 6213,  label: "TrueCrypt RIPEMD160 + XTS 1536 bit | FDE" },
        { value: 6241,  label: "TrueCrypt RIPEMD160 + XTS 512 bit + boot-mode | FDE" },
        { value: 6242,  label: "TrueCrypt RIPEMD160 + XTS 1024 bit + boot-mode | FDE" },
        { value: 6243,  label: "TrueCrypt RIPEMD160 + XTS 1536 bit + boot-mode | FDE" },
        { value: 6221,  label: "TrueCrypt SHA512 + XTS 512 bit | FDE" },
        { value: 6222,  label: "TrueCrypt SHA512 + XTS 1024 bit | FDE" },
        { value: 6223,  label: "TrueCrypt SHA512 + XTS 1536 bit | FDE" },
        { value: 6231,  label: "TrueCrypt Whirlpool + XTS 512 bit | FDE" },
        { value: 6232,  label: "TrueCrypt Whirlpool + XTS 1024 bit | FDE" },
        { value: 6233,  label: "TrueCrypt Whirlpool + XTS 1536 bit | FDE" },

        // Documents
        { value: 10400, label: "PDF 1.1 - 1.3 (Acrobat 2 - 4) | Documents" },
        { value: 10410, label: "PDF 1.1 - 1.3 (Acrobat 2 - 4), collider #1 | Documents" },
        { value: 10420, label: "PDF 1.1 - 1.3 (Acrobat 2 - 4), collider #2 | Documents" },
        { value: 10500, label: "PDF 1.4 - 1.6 (Acrobat 5 - 8) | Documents" },
        { value: 10600, label: "PDF 1.7 Level 3 (Acrobat 9) | Documents" },
        { value: 10700, label: "PDF 1.7 Level 8 (Acrobat 10 - 11) | Documents" },
        { value: 9400,  label: "MS Office 2007 | Documents" },
        { value: 9500,  label: "MS Office 2010 | Documents" },
        { value: 9600,  label: "MS Office 2013 | Documents" },
        { value: 9700,  label: "MS Office <= 2003 $0/$1, MD5 + RC4 | Documents" },
        { value: 9710,  label: "MS Office <= 2003 $0/$1, MD5 + RC4, collider #1 | Documents" },
        { value: 9720,  label: "MS Office <= 2003 $0/$1, MD5 + RC4, collider #2 | Documents" },
        { value: 9800,  label: "MS Office <= 2003 $3/$4, SHA1 + RC4 | Documents" },
        { value: 9810,  label: "MS Office <= 2003 $3, SHA1 + RC4, collider #1 | Documents" },
        { value: 9820,  label: "MS Office <= 2003 $3, SHA1 + RC4, collider #2 | Documents" },
        { value: 18400, label: "Open Document Format (ODF) 1.2 (SHA-256, AES) | Documents" },
        { value: 18600, label: "Open Document Format (ODF) 1.1 (SHA-1, Blowfish) | Documents" },
        { value: 16200, label: "Apple Secure Notes | Documents" },

        // Password Managers
        { value: 15500, label: "JKS Java Key Store Private Keys (SHA1) | Password Managers" },
        { value: 6600,  label: "1Password, agilekeychain | Password Managers" },
        { value: 8200,  label: "1Password, cloudkeychain | Password Managers" },
        { value: 9000,  label: "Password Safe v2 | Password Managers" },
        { value: 5200,  label: "Password Safe v3 | Password Managers" },
        { value: 6800,  label: "LastPass + LastPass sniffed | Password Managers" },
        { value: 13400, label: "KeePass 1 (AES/Twofish) & KeePass 2 (AES) | Password Managers" },
        { value: 11300, label: "Bitcoin/Litecoin wallet.dat | Password Managers" },
        { value: 16600, label: "Electrum Wallet (Salt-Type 1-3) | Password Managers" },
        { value: 21700, label: "Electrum Wallet (Salt-Type 4) | Password Managers" },
        { value: 21800, label: "Electrum Wallet (Salt-Type 5) | Password Managers" },
        { value: 12700, label: "Blockchain, My Wallet | Password Managers" },
        { value: 15200, label: "Blockchain, My Wallet, V2 | Password Managers" },
        { value: 18800, label: "Blockchain, My Wallet, Second Password (SHA256) | Password Managers" },
        { value: 23100, label: "Apple Keychain | Password Managers" },
        { value: 16300, label: "Ethereum Pre-Sale Wallet, PBKDF2-HMAC-SHA256 | Password Managers" },
        { value: 15600, label: "Ethereum Wallet, PBKDF2-HMAC-SHA256 | Password Managers" },
        { value: 15700, label: "Ethereum Wallet, SCRYPT | Password Managers" },
        { value: 22500, label: "MultiBit Classic .key (MD5) | Password Managers" },
        { value: 22700, label: "MultiBit HD (scrypt) | Password Managers" },

        // Archives
        { value: 11600, label: "7-Zip | Archives" },
        { value: 12500, label: "RAR3-hp | Archives" },
        { value: 13000, label: "RAR5 | Archives" },
        { value: 17200, label: "PKZIP (Compressed) | Archives" },
        { value: 17220, label: "PKZIP (Compressed Multi-File) | Archives" },
        { value: 17225, label: "PKZIP (Mixed Multi-File) | Archives" },
        { value: 17230, label: "PKZIP (Mixed Multi-File Checksum-Only) | Archives" },
        { value: 17210, label: "PKZIP (Uncompressed) | Archives" },
        { value: 20500, label: "PKZIP Master Key | Archives" },
        { value: 20510, label: "PKZIP Master Key (6 byte optimization) | Archives" },
        { value: 14700, label: "iTunes backup < 10.0 | Archives" },
        { value: 14800, label: "iTunes backup >= 10.0 | Archives" },
        { value: 23001, label: "SecureZIP AES-128 | Archives" },
        { value: 23002, label: "SecureZIP AES-192 | Archives" },
        { value: 23003, label: "SecureZIP AES-256 | Archives" },
        { value: 13600, label: "WinZip | Archives" },
        { value: 18900, label: "Android Backup | Archives" },
        { value: 13200, label: "AxCrypt | Archives" },
        { value: 13300, label: "AxCrypt in-memory SHA1 | Archives" },

        // Forums, CMS, E-Commerce
        { value: 8400,  label: "WBB3 (Woltlab Burning Board)" },
        { value: 2611,  label: "vBulletin < v3.8.5" },
        { value: 2711,  label: "vBulletin >= v3.8.5" },
        { value: 2612,  label: "PHPS" },
        { value: 121,   label: "SMF (Simple Machines Forum) > v1.1" },
        { value: 3711,  label: "MediaWiki B type" },
        { value: 4521,  label: "Redmine" },
        { value: 11,    label: "Joomla < 2.5.18" },
        { value: 13900, label: "OpenCart" },
        { value: 11000, label: "PrestaShop" },
        { value: 16000, label: "Tripcode" },
        { value: 7900,  label: "Drupal7" },
        { value: 21,    label: "osCommerce, xt:Commerce" },
        { value: 4522,  label: "PunBB" },
        { value: 2811,  label: "MyBB 1.2+, IPB2+ (Invision Power Board)" },

        // One-Time Passwords
        { value: 18100, label: "TOTP (HMAC-SHA1)" },

        // Plaintext
        { value: 2000,  label: "STDOUT" },
        { value: 99999, label: "Plaintext" },

        // Framework
        { value: 21600, label: "Web2py pbkdf2-sha512" },
        { value: 10000, label: "Django (PBKDF2-SHA256)" },
        { value: 124,   label: "Django (SHA-1)" },
    ];

    populateSelect("#attackMode", attackModes, "Select Attack Mode");
    populateSelect("#hashMode", hashModes, "Select Hash Mode");

    // Enable/disable wordlist based on selected attack mode
    $("#attackMode").change(function () {
        const selectedMode = parseInt($(this).val(), 10);
        $("#wordlist").prop("disabled", ![0, 1, 6, 7].includes(selectedMode));
    });

    // Populate the client select if global clientUUIDs exists
    if (typeof clientUUIDs === "string" && $("#clientUUID").length) {
        const $clientSelect = $("#clientUUID").empty().append(
            '<option value="">-- Select a client --</option>'
        );
        clientUUIDs.split(";").forEach((uuid) => {
            const [name, realUUID] = uuid.split(":");
            if (realUUID && realUUID.trim()) {
                $clientSelect.append(
                    `<option value="${realUUID.trim()}">${name.trim()}</option>`
                );
            }
        });
    }

    // ------------------------------------------------
    // MODAL & ACTION BUTTONS (HANDSHAKE)
    // ------------------------------------------------
    $(document).on("click", ".crack-btn", function () {
        const uuid = $(this).data("uuid");
        $("#crackUUID").val(uuid);
        $("#crackModal").modal("show");
    });

    $(document).on("click", ".delete-btn", function () {
        const uuid = $(this).data("uuid");
        $("#deleteUUID").val(uuid);
        $("#deleteConfirmModal").modal("show");
    });

    $(document).on("click", ".hashcat-options-btn", function () {
        const options = $(this).data("options");
        $("#hashcatOptionsContent").text(options !== "<nil>" ? options : "No scan run");
        $("#hashcatOptionsModal").modal("show");
    });

    $(document).on("click", ".hashcat-logs-btn", function () {
        const logs = $(this).data("logs");
        // Note: using .html here; ensure logs are sanitized to avoid XSS
        $("#hashcatLogsContent").html(logs.replace(/\n/g, "<br>") || "No scan run");
        $("#hashcatLogsModal").modal("show");
    });

    // ------------------------------------------------
    // TABLE SEARCH
    // ------------------------------------------------
    $("#searchInput").on("keyup", function () {
        const searchTerm = $(this).val().toLowerCase();
        ["#handshakeTableBody", "#clientTableBody", "#raspberrypiTableBody"].forEach(
            (selector) => {
                $(selector)
                    .find("tr")
                    .each(function () {
                        $(this).toggle($(this).text().toLowerCase().includes(searchTerm));
                    });
            }
        );
    });

    // ------------------------------------------------
    // TOOLTIP INITIALIZATION
    // ------------------------------------------------
    $('[data-toggle="tooltip"]').tooltip();

    // Enable the wordlist by default (Straight mode)
    if ($("#wordlist").length) {
        $("#wordlist").prop("disabled", false);
    }

    // ------------------------------------------------
    // PAGINATION WITH AJAX (HANDSHAKE)
    // ------------------------------------------------
    $(document).on("click", ".pagination .page-link", function (e) {
        e.preventDefault();
        const href = $(this).attr("href");
        if (href) {
            showLoading();
            $("#page-content-wrapper").removeClass("loaded");
            const currentDarkMode = isDarkMode();

            $.ajax({
                url: href,
                success: function (data) {
                    const $newContent = $(data).find("#page-content-wrapper").html();
                    $("#page-content-wrapper").html($newContent);
                    setDarkMode(currentDarkMode);
                    setTimeout(() => {
                        $("#page-content-wrapper").addClass("loaded");
                        hideLoading();
                    }, 100);
                },
                error: function () {
                    hideLoading();
                    alert("Failed to load page.");
                },
            });
        }
    });

    // ------------------------------------------------
    // URL PARAMETER: DARK MODE
    // ------------------------------------------------
    const urlParams = new URLSearchParams(window.location.search);
    const darkModeParam = urlParams.get("darkMode");
    if (darkModeParam !== null) {
        setDarkMode(darkModeParam === "true");
    }

    // ------------------------------------------------
    // ENCRYPTION DETAILS & COPY BUTTONS
    // ------------------------------------------------
    const showEncryptionDetails = (caCert, clientCert, clientKey) => {
        $("#caCert").val(caCert);
        $("#clientCert").val(clientCert);
        $("#clientKey").val(clientKey);
        $("#encryptionDetailsModal").modal("show");
    };

    $(document).on("click", ".copy-btn", function () {
        const targetId = $(this).data("target");
        const textArea = document.getElementById(targetId);
        textArea.select();
        document.execCommand("copy");

        const $btn = $(this);
        const originalText = $btn.text();
        $btn.text("Copied!");
        setTimeout(() => $btn.text(originalText), 1500);
    });

    $(document).on("click", ".show-certs-btn", function () {
        const caCert = $(this).data("ca-cert");
        const clientCert = $(this).data("client-cert");
        const clientKey = $(this).data("client-key");
        showEncryptionDetails(caCert, clientCert, clientKey);
    });

    // Toggle encryption state and update associated form field
    $(document).on("change", ".encryption-toggle", function () {
        const isEnabled = this.checked;

        // Find the associated hidden input and update its value
        $(this).closest("form").find('input[name="enabled"]').val(isEnabled ? "true" : "false");

        // Submit the form automatically
        this.form.submit();
    });
});
