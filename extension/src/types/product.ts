// product.ts - Product information types

export interface ProductInfo {
  name: string;
  brand?: string;
  size?: string;
  upc?: string;
  url: string;
  retailer: 'walmart';
}

export interface SearchRequest {
  productName: string;
  brand?: string;
  size?: string;
}
