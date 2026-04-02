/**
 * 递归将对象/数组中所有 key 从 snake_case 转为 camelCase
 */
export function toCamelCase(obj: any): any {
  if (obj === null || obj === undefined) return obj;
  if (Array.isArray(obj)) return obj.map(toCamelCase);
  if (obj instanceof Date || obj instanceof Blob || obj instanceof File) return obj;
  if (typeof obj === 'object') {
    return Object.keys(obj).reduce((acc: any, key: string) => {
      const camelKey = key.replace(/_([a-z0-9])/g, (_, c) => c.toUpperCase());
      acc[camelKey] = toCamelCase(obj[key]);
      return acc;
    }, {});
  }
  return obj;
}

/**
 * 递归将对象/数组中所有 key 从 camelCase 转为 snake_case
 */
export function toSnakeCase(obj: any): any {
  if (obj === null || obj === undefined) return obj;
  if (Array.isArray(obj)) return obj.map(toSnakeCase);
  if (obj instanceof Date || obj instanceof Blob || obj instanceof File || obj instanceof FormData) return obj;
  if (typeof obj === 'object') {
    return Object.keys(obj).reduce((acc: any, key: string) => {
      const snakeKey = key
        .replace(/([A-Z]+)([A-Z][a-z])/g, '$1_$2')
        .replace(/([a-z0-9])([A-Z])/g, '$1_$2')
        .toLowerCase();
      acc[snakeKey] = toSnakeCase(obj[key]);
      return acc;
    }, {});
  }
  return obj;
}
