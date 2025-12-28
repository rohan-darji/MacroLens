// messages.ts - Message passing types for Chrome extension communication

import { NutritionData } from './nutrition';
import { ProductInfo } from './product';

export enum MessageType {
  GET_NUTRITION = 'GET_NUTRITION',
  NUTRITION_RESPONSE = 'NUTRITION_RESPONSE',
  CLEAR_CACHE = 'CLEAR_CACHE',
  CACHE_CLEARED = 'CACHE_CLEARED',
}

export interface Message<T = any> {
  type: MessageType;
  payload: T;
}

export interface GetNutritionPayload {
  productInfo: ProductInfo;
  timestamp: number;
}

export interface NutritionResponsePayload {
  data?: NutritionData;
  error?: string;
  cached: boolean;
}
