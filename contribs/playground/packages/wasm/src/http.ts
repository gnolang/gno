/// <reference lib="webworker" />

// RequestHandlerFunc defines the function type for fetch request handlers.
type RequestHandlerFunc = (r: Request) => Promise<Response>

// ServiceWorkerNS defines the namespace necessary for registering
// a service worker fetch handler.
interface ServiceWorkerNS {
  readonly worker: FetchServiceWorker

  registerFetchHandler: (h: RequestHandlerFunc) => void
  fetchHandler: () => RequestHandlerFunc | null
}

// ServiceWorkerWasmGlobalScope defines the global scope
// for the service worker that runs the wasm instance.
export interface ServiceWorkerWasmGlobalScope extends ServiceWorkerGlobalScope {
  wasm: ServiceWorkerNS
}

// FetchServiceWorker simplifies running a service worker that in turn runs a wasm instance.
// It does so by implementing a basic service worker that installs and activates automatically
// and response to requests by calling a handler function.
class FetchServiceWorker {
  public readonly worker: ServiceWorkerGlobalScope

  public handler: RequestHandlerFunc | null = null
  public allow: Array<string | RegExp> = []
  public ignore: Array<string | RegExp> = []

  constructor(worker: ServiceWorkerGlobalScope) {
    this.worker = worker
    this.setupEventListeners()
  }

  private setupEventListeners() {
    this.worker.addEventListener('install', this.handleWorkerInstall.bind(this))
    this.worker.addEventListener('activate', this.handleWorkerActivate.bind(this))
    this.worker.addEventListener('fetch', this.handleFetch.bind(this))
  }

  private handleWorkerInstall(evt: ExtendableEvent) {
    evt.waitUntil(this.worker.skipWaiting())
  }

  private handleWorkerActivate(evt: ExtendableEvent) {
    evt.waitUntil(this.worker.clients.claim())
  }

  private handleFetch(evt: FetchEvent) {
    if (!this.handler) {
      console.error('No fetch handler is registered')
      return
    }

    if (this.isAllowed(evt.request.url)) {
      evt.respondWith(this.handler(evt.request))
    }
  }

  isAllowed(url: string): boolean {
    const { pathname } = new URL(url)

    if (pathname === '/') {
      return false
    }

    // First check for ignored URL paths
    const isIgnored = this.ignore.some((exp) => {
      return new RegExp(exp).test(pathname)
    })
    if (isIgnored) {
      return false
    }

    // If there are allowed paths make sure that current path is one of them
    if (this.allow) {
      return this.allow.some((exp) => {
        return new RegExp(exp).test(pathname)
      })
    }

    return true
  }
}

// Rules defines service worker URL path handling rules.
interface Rules {
  allow?: Array<string | RegExp>
  ignore?: Array<string | RegExp>
}

// Create a service worker wasm namespace.
//
// The namespace has support for registering a fetch request handler
// that is called for all application requests.
// The handler acts like a middleware between the main thread application
// and the remote servers.
//
// Usage:
//
//   createServiceWorkerNS(self, {
//     ignore: ['/favicon.ico'],
//   })
//
// The registration is done by calling `wasm.registerFetchHandler(handler)`
// which should be done from within the WebAssembly instance. All requests
// will be forwarded to the handler once registered.
export function createServiceWorkerNS(scope: ServiceWorkerWasmGlobalScope, rules?: Rules): ServiceWorkerNS {
  const worker = new FetchServiceWorker(scope)

  if (rules?.allow) {
    worker.allow = rules.allow
  }

  if (rules?.ignore) {
    worker.ignore = rules.ignore
  }

  function fetchHandler(): RequestHandlerFunc | null {
    return worker.handler
  }

  function registerFetchHandler(h: RequestHandlerFunc) {
    worker.handler = h
  }

  const ns: ServiceWorkerNS = {
    worker,
    fetchHandler,
    registerFetchHandler,
  }

  // Make namespace available as a service worker global
  scope.wasm = ns

  return ns
}
