export * from "./user";
export * from "./common";
export * from "./client";
export * from "./dashboard";
export * from "./version";
export * from "./ops";
export * from "./workspace";

export type BaseResponse<T> = {
  code: number;
  msg: string;
  data: T | null;
};

export type PageResponse<T> = {
  total: number;
  list: T[];
};