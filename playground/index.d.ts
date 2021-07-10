interface ActionlintError {
    kind: string;
    message: string;
    line: number;
    column: number;
}

interface Window {
    runActionlint?(src: string): void;
    getYamlSource(): string;
    showError(msg: string): void;
    onCheckCompleted(errs: ActionlintError[]): void;
    dismissLoading(): void;
}

declare class Go {
    importObject: Imports;
    run(mod: Instance): Promise<any>;
}

declare namespace IsMobile {
    import type { isMobileResult } from 'ismobilejs';
}
declare const isMobile: IsMobile.isMobileResult;
