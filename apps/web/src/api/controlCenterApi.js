const defaultRequestTimeoutMs = 8000;

function createRequestController(timeoutMs, externalSignal) {
  const controller = new AbortController();
  let timeoutId = null;
  let didTimeout = false;

  const abortFromExternalSignal = () => {
    controller.abort(externalSignal?.reason);
  };

  if (externalSignal?.aborted) {
    abortFromExternalSignal();
  } else {
    externalSignal?.addEventListener("abort", abortFromExternalSignal, { once: true });
  }

  if (timeoutMs > 0) {
    timeoutId = globalThis.setTimeout(() => {
      didTimeout = true;
      controller.abort();
    }, timeoutMs);
  }

  return {
    didTimeout: () => didTimeout,
    signal: controller.signal,
    cleanup() {
      if (timeoutId) {
        globalThis.clearTimeout(timeoutId);
      }
      externalSignal?.removeEventListener("abort", abortFromExternalSignal);
    }
  };
}

async function readJsonPayload(response) {
  const text = await response.text();
  if (!text) {
    return null;
  }
  try {
    return JSON.parse(text);
  } catch {
    if (response.ok) {
      throw new Error(`invalid JSON response: ${response.status}`);
    }
    return null;
  }
}

export async function requestJson(path, options = {}) {
  const {
    headers,
    signal: externalSignal,
    timeoutMs = defaultRequestTimeoutMs,
    ...fetchOptions
  } = options;
  const requestController = createRequestController(timeoutMs, externalSignal);

  try {
    const response = await fetch(path, {
      headers: {
        "Content-Type": "application/json",
        ...(headers ?? {})
      },
      ...fetchOptions,
      signal: requestController.signal
    });
    const payload = await readJsonPayload(response);
    if (!response.ok) {
      throw new Error(payload?.error ?? `request failed: ${response.status}`);
    }
    return payload ?? {};
  } catch (error) {
    if (requestController.didTimeout()) {
      throw new Error(`request timed out after ${timeoutMs}ms`);
    }
    throw error;
  } finally {
    requestController.cleanup();
  }
}

export function websocketUrlWithQuery(baseUrl, params) {
  const locationHref = globalThis.window?.location?.href ?? "http://127.0.0.1/";
  const url = new URL(baseUrl, locationHref);
  Object.entries(params).forEach(([key, value]) => {
    url.searchParams.set(key, value);
  });
  return url.toString();
}
