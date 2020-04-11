import { CMS } from "netlify-cms-core";
import {
  Implementation as CMSBackendImpl,
  ImplementationEntry,
} from "netlify-cms-lib-util/src/implementation";
import ReactMod from "react";

interface GoGlobal {
  new (): Go;
}

interface Go {
  run(instance: WebAssembly.Instance): Promise<void>;
  importObject: Record<string, Record<string, WebAssembly.ImportValue>>;
}

interface FileInfo {
  name: string;
  size: number;
  isDir: boolean;
}

interface BackendFs {
  stat(path: string, cb: (err: string | null, meta?: FileInfo) => void): void;
  readFile(
    path: string,
    cb: (err: string | null, file?: Uint8Array) => void
  ): void;
  listDir(
    path: string,
    cb: (err: string | null, files?: FileInfo[]) => void
  ): void;
}

interface HugoPreview {
  setBackendFs(fs: BackendFs): void;
}

declare global {
  interface Window {
    CMS: CMS;
    Go: GoGlobal;
    HugoPreview: HugoPreview;
    React: typeof ReactMod;
  }
}

const go = new window.Go();
let wasm_init: Promise<void> | null = null;
const initWasm = async () => {
  // todo: polyfill
  const { instance } = await WebAssembly.instantiateStreaming(
    fetch("/dist/previewer.wasm"),
    go.importObject
  );
  go.run(instance);
};

class CMSBackendFs implements BackendFs {
  backend: CMSBackendImpl;
  constructor(backend: CMSBackendImpl) {
    this.backend = backend;
    console.log("making backendFS", { backend });
  }

  private static entryToFileInfo(
    entry: ImplementationEntry,
    isDir = false
  ): FileInfo {
    return {
      isDir,
      name: entry.file.path.split("/").pop() || "",
      size: entry.data.length,
    };
  }

  stat(path: string, cb: (err: string | null, meta?: FileInfo) => void) {
    this.backend.getEntry(path).then(
      (val) => cb(null, CMSBackendFs.entryToFileInfo(val)),
      (err) => cb(err)
    );
  }

  readFile(path: string, cb: (err: string | null, file?: Uint8Array) => void) {
    this.backend.getEntry(path).then(
      (val) => cb(null, new TextEncoder().encode(val.data)),
      (err) => cb(err)
    );
  }

  listDir(path: string, cb: (err: string | null, files?: FileInfo[]) => void) {
    this.backend
      .entriesByFolder(
        path,
        ".html" /* todo: allow to get rid of this */,
        1 /* todo: unused in github, i think */
      )
      .then(
        (val) => {
          const files = val.map((entry) => CMSBackendFs.entryToFileInfo(entry));
          cb(null, files);
        },
        (err) => cb(err)
      );
  }
}

class Renderer {
  hugo: HugoPreview;
  constructor(backend: CMSBackendImpl) {
    this.hugo = window.HugoPreview;
    const backendFS = new CMSBackendFs(backend);
    this.hugo.setBackendFs(backendFS);
  }
}

const initRenderer = async (): Promise<Renderer> => {
  if (wasm_init === null) {
    wasm_init = initWasm();
  }
  await wasm_init;

  // todo: Fix typing in CMS
  const backend: CMSBackendImpl | undefined = window.CMS.getBackend(
    "proxy"
  )?.init({
    backend: {
      name: "proxy",
      proxy_url: "http://localhost:8081/api/v1",
    },
    local_backend: true,
  });
  if (!backend) {
    throw new Error("Backend not found");
  }
  return new Renderer(backend);
};

const { useEffect, useState } = window.React;
const React = window.React;

class PostPreview extends React.Component<{ entry: any }> {
  componentDidMount() {
    initRenderer(); // todo: set result in state
    console.log({ props: this.props });
  }

  componentDidUpdate() {
    console.log(this.props.entry);
  }

  render() {
    return <>Hello World!</>;
  }
}

window.CMS.registerPreviewTemplate("post", PostPreview as any); // todo: fix typing in CMS
