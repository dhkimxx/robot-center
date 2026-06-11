const listQueryKeys = ["limit", "offset", "sort", "order", "filter"];

export function createListQueryPath(path, query = {}) {
  const searchParams = new URLSearchParams();
  listQueryKeys.forEach((key) => {
    const value = query[key];
    if (value === null || value === undefined || value === "") {
      return;
    }
    searchParams.set(key, String(value));
  });
  const queryString = searchParams.toString();
  return queryString ? `${path}?${queryString}` : path;
}
