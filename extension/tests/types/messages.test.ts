import { describe, it, expect } from 'vitest';
import { MessageType } from '@/types/messages';
import type { Message, GetNutritionPayload, NutritionResponsePayload } from '@/types/messages';

describe('Message Types', () => {
  describe('MessageType enum', () => {
    it('should have GET_NUTRITION type', () => {
      expect(MessageType.GET_NUTRITION).toBe('GET_NUTRITION');
    });

    it('should have NUTRITION_RESPONSE type', () => {
      expect(MessageType.NUTRITION_RESPONSE).toBe('NUTRITION_RESPONSE');
    });

    it('should have CLEAR_CACHE type', () => {
      expect(MessageType.CLEAR_CACHE).toBe('CLEAR_CACHE');
    });

    it('should have CACHE_CLEARED type', () => {
      expect(MessageType.CACHE_CLEARED).toBe('CACHE_CLEARED');
    });
  });

  describe('Message structure', () => {
    it('should create valid GET_NUTRITION message', () => {
      const message: Message<GetNutritionPayload> = {
        type: MessageType.GET_NUTRITION,
        payload: {
          productInfo: {
            name: 'Whole Milk',
            brand: 'Great Value',
            url: 'https://www.walmart.com/ip/test/123',
            retailer: 'walmart',
          },
          timestamp: Date.now(),
        },
      };

      expect(message.type).toBe(MessageType.GET_NUTRITION);
      expect(message.payload.productInfo.name).toBe('Whole Milk');
      expect(message.payload.timestamp).toBeGreaterThan(0);
    });

    it('should create valid NUTRITION_RESPONSE message with data', () => {
      const message: Message<NutritionResponsePayload> = {
        type: MessageType.NUTRITION_RESPONSE,
        payload: {
          data: {
            fdcId: '12345',
            productName: 'Whole Milk',
            servingSize: '100',
            servingSizeUnit: 'g',
            nutrients: {
              calories: 149,
              protein: 7.7,
              carbohydrates: 11.7,
              totalFat: 7.9,
            },
            confidence: 92.5,
            source: 'USDA',
          },
          cached: false,
        },
      };

      expect(message.type).toBe(MessageType.NUTRITION_RESPONSE);
      expect(message.payload.data?.fdcId).toBe('12345');
      expect(message.payload.cached).toBe(false);
    });

    it('should create valid NUTRITION_RESPONSE message with error', () => {
      const message: Message<NutritionResponsePayload> = {
        type: MessageType.NUTRITION_RESPONSE,
        payload: {
          error: 'Product not found',
          cached: false,
        },
      };

      expect(message.type).toBe(MessageType.NUTRITION_RESPONSE);
      expect(message.payload.error).toBe('Product not found');
      expect(message.payload.data).toBeUndefined();
    });

    it('should create valid CLEAR_CACHE message', () => {
      const message: Message = {
        type: MessageType.CLEAR_CACHE,
        payload: {},
      };

      expect(message.type).toBe(MessageType.CLEAR_CACHE);
    });
  });

  describe('Payload validation', () => {
    it('should have required fields in GetNutritionPayload', () => {
      const payload: GetNutritionPayload = {
        productInfo: {
          name: 'Test Product',
          url: 'https://www.walmart.com/ip/test/123',
          retailer: 'walmart',
        },
        timestamp: Date.now(),
      };

      expect(payload.productInfo.name).toBeDefined();
      expect(payload.productInfo.url).toBeDefined();
      expect(payload.productInfo.retailer).toBe('walmart');
      expect(payload.timestamp).toBeGreaterThan(0);
    });

    it('should allow optional fields in ProductInfo', () => {
      const payload: GetNutritionPayload = {
        productInfo: {
          name: 'Test Product',
          brand: 'Test Brand',
          size: '1 gallon',
          upc: '123456789',
          url: 'https://www.walmart.com/ip/test/123',
          retailer: 'walmart',
        },
        timestamp: Date.now(),
      };

      expect(payload.productInfo.brand).toBe('Test Brand');
      expect(payload.productInfo.size).toBe('1 gallon');
      expect(payload.productInfo.upc).toBe('123456789');
    });
  });
});
