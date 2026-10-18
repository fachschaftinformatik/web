// Browser module vendored by Nix. Declare only the options used by this app.
type CountMessage = { one: string; other: string };

export interface FrontendOptions {
  scope?: ParentNode;
  fileUpload?: {
    i18n: {
      chooseFilesButton: string;
      dropInstruction: string;
      noFileChosen: string;
      multipleFilesChosen: CountMessage;
      enteredDropZone: string;
      leftDropZone: string;
    };
  };
  passwordInput?: {
    i18n: {
      showPassword: string;
      hidePassword: string;
      showPasswordAriaLabel: string;
      hidePasswordAriaLabel: string;
      passwordShownAnnouncement: string;
      passwordHiddenAnnouncement: string;
    };
  };
}

export function initAll(options?: FrontendOptions): void;
