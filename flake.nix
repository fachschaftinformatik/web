{
  description = "Development environment and browser assets for the Fachschaft website";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }:
    let
      forAllSystems = nixpkgs.lib.genAttrs [
        "aarch64-darwin"
        "x86_64-darwin"
        "aarch64-linux"
        "x86_64-linux"
      ];
    in
    {
      packages = forAllSystems (system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          govukVersion = "6.6.0";
          htmxVersion = "2.0.8";
          tinymceVersion = "8.10.0";
          languageVersion = "26.9.28";
          iconsRevision = "c053cf1c198580723c8768f938cb6146dcd122e8";
          communityRevision = "a1bfd60d1e0fcc2de52a1d357b934c3098ec62a6";

          govuk = pkgs.fetchurl {
            url = "https://registry.npmjs.org/govuk-frontend/-/govuk-frontend-${govukVersion}.tgz";
            hash = "sha256-UsTrVU5U2KbkuoAenaXsOluTD9LxF6OWd/C2KW6VR1E=";
          };
          govukLicense = pkgs.fetchurl {
            url = "https://raw.githubusercontent.com/alphagov/govuk-frontend/v${govukVersion}/LICENSE.txt";
            hash = "sha256-xlLzWx8uW4jrSfHoqTXux5tHYagWZqzlduGvcj9Kkjg=";
          };
          htmx = pkgs.fetchurl {
            url = "https://registry.npmjs.org/htmx.org/-/htmx.org-${htmxVersion}.tgz";
            hash = "sha256-UTUbKsSzvAZWkqTdE0moL5C/fFJhNJN8yyXTkyyM6VQ=";
          };
          tinymce = pkgs.fetchurl {
            url = "https://registry.npmjs.org/tinymce/-/tinymce-${tinymceVersion}.tgz";
            hash = "sha256-yCfom0e+9PAuKJLaM19UCSU1n3VGeQWOajDLsEmUba4=";
          };
          language = pkgs.fetchurl {
            url = "https://registry.npmjs.org/tinymce-i18n/-/tinymce-i18n-${languageVersion}.tgz";
            hash = "sha256-rZc4XVb75ZIoQDyWm5C/4R1/LHtlLhrO8T3+pGsUlgc=";
          };
          icons = pkgs.fetchurl {
            name = "govspeak-visual-editor-${iconsRevision}.tar.gz";
            url = "https://codeload.github.com/alphagov/govspeak-visual-editor/tar.gz/${iconsRevision}";
            hash = "sha256-vzZthKLNKaXfeaiXDxts0YK8WyBO+ucNJnDADUlacdQ=";
          };
          community = pkgs.fetchurl {
            name = "govuk-design-system-${communityRevision}.tar.gz";
            url = "https://codeload.github.com/alphagov/govuk-design-system/tar.gz/${communityRevision}";
            hash = "sha256-OIvJjeUvtn4dAkpNzAzLwEXtQzApwz+vaGATLr9wgDU=";
          };
          assets = pkgs.stdenvNoCC.mkDerivation {
            pname = "web-browser-assets";
            version = "1";
            dontUnpack = true;
            nativeBuildInputs = [ pkgs.dart-sass ];

            buildPhase = ''
              runHook preBuild
              for name in govuk htmx tinymce language icons community; do
                mkdir -p "$name"
              done
              tar -xzf ${govuk} -C govuk --strip-components=1
              tar -xzf ${htmx} -C htmx --strip-components=1
              tar -xzf ${tinymce} -C tinymce --strip-components=1
              tar -xzf ${language} -C language --strip-components=1
              tar -xzf ${icons} -C icons --strip-components=1
              tar -xzf ${community} -C community --strip-components=1

              sass --no-source-map --style=compressed --quiet-deps \
                --load-path=govuk/dist ${./view/styles/govuk.scss} govuk.css
              sass --no-source-map --style=compressed ${./view/styles/site.scss} site.css
              sass --no-source-map --style=compressed ${./view/styles/editor.scss} editor.css
              runHook postBuild
            '';

            installPhase = ''
              runHook preInstall
              mkdir -p "$out/fonts" "$out/icons" "$out/community"
              install -m 0644 govuk.css "$out/govuk.css"
              install -m 0644 site.css editor.css "$out/"
              install -m 0644 govuk/dist/govuk/govuk-frontend.min.js "$out/govuk.js"
              install -m 0644 govuk/dist/govuk/assets/fonts/*.woff2 "$out/fonts/"
              install -m 0644 ${govukLicense} "$out/GOVUK-LICENSE.txt"
              install -m 0644 htmx/dist/htmx.min.js "$out/htmx.js"
              install -m 0644 htmx/LICENSE "$out/HTMX-LICENSE.txt"

              for path in tinymce.min.js themes/silver/theme.min.js \
                icons/default/icons.min.js models/dom/model.min.js \
                skins/content/default/content.min.css plugins/lists/plugin.min.js; do
                mkdir -p "$out/tinymce/$(dirname "$path")"
                install -m 0644 "tinymce/$path" "$out/tinymce/$path"
              done
              mkdir -p "$out/tinymce/skins/ui/oxide" "$out/tinymce/langs8"
              install -m 0644 tinymce/skins/ui/oxide/*.min.css "$out/tinymce/skins/ui/oxide/"
              install -m 0644 language/langs8/de.js "$out/tinymce/langs8/de.js"
              install -m 0644 tinymce/license.md "$out/tinymce/LICENSE.txt"
              install -m 0644 language/LICENSE.md "$out/tinymce/I18N-LICENSE.md"

              for name in bullet-list ordered-list link undo redo; do
                install -m 0644 "icons/lib/icons/$name.svg" "$out/icons/$name.svg"
              done
              install -m 0644 icons/LICENCE "$out/icons/GDS-LICENCE.txt"
              for name in discuss-and-give-feedback.svg share-your-examples-and-research.svg \
                propose-a-change-to-pages.svg research-interview.png catchup-groupshot.png workshop.png; do
                install -m 0644 "community/src/community/images/$name" "$out/community/$name"
              done
              install -m 0644 community/LICENSE "$out/community/GDS-LICENCE.txt"
              runHook postInstall
            '';
          };
        in
        {
          inherit assets;
          default = assets;
        });

      apps = forAllSystems (system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          vendor = pkgs.writeShellApplication {
            name = "vendor-assets";
            runtimeInputs = [ pkgs.coreutils ];
            text = ''
              root="''${1:-$PWD}"
              if [ ! -f "$root/view/styles/govuk.scss" ]; then
                echo "Run from the repository root or pass its path." >&2
                exit 1
              fi
              destination="$root/view/assets"
              mkdir -p "$destination"
              rm -rf "$destination/fonts" "$destination/icons" "$destination/community" "$destination/tinymce"
              cp -R --no-preserve=mode,ownership ${self.packages.${system}.assets}/. "$destination/"
            '';
          };
        in
        {
          vendor = {
            type = "app";
            program = "${vendor}/bin/vendor-assets";
          };
        });

      checks = forAllSystems (system: {
        assets = self.packages.${system}.assets;
      });

      devShells = forAllSystems (system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          goVersion = builtins.head (builtins.match ".*\ngo ([0-9.]+)\n.*" (builtins.readFile ./go.mod));
          go = pkgs.${"go_" + builtins.replaceStrings [ "." ] [ "_" ] (pkgs.lib.versions.majorMinor goVersion)};
        in
        {
          default = assert pkgs.lib.assertMsg (go.version == goVersion)
            "The locked Nix Go version must match go.mod; update go.mod and flake.lock together.";
            pkgs.mkShellNoCC {
              packages = [ go ];
              GOTOOLCHAIN = "local";
              shellHook = ''
                ${self.apps.${system}.vendor.program} "$PWD"
                export ACCOUNT_DOMAINS="''${ACCOUNT_DOMAINS-studmail.w-hs.de,fsv-wh.de}"
                export REGISTRATION_DOMAINS="''${REGISTRATION_DOMAINS-studmail.w-hs.de}"
              '';
            };
        });
    };
}
