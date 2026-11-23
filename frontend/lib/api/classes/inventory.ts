import { BaseAPI, route } from "../base";
import type {
  InventoryAssignmentOut,
  InventoryIssueInput,
  InventoryRecipientCreate,
  InventoryRecipientOut,
  InventoryRecipientUpdate,
  InventoryReturnInput,
  InventoryStockAddInput,
  InventoryStockWriteOffInput,
  InventoryTransactionOut,
} from "../types/data-contracts";

export class InventoryApi extends BaseAPI {
  recipients(params: { search?: string; active?: boolean; page?: number; pageSize?: number } = {}) {
    return this.http.get<InventoryRecipientOut[]>({ url: route("/inventory/recipients", params) });
  }

  recipient(id: string) {
    return this.http.get<InventoryRecipientOut>({ url: route(`/inventory/recipients/${id}`) });
  }

  createRecipient(body: InventoryRecipientCreate) {
    return this.http.post<InventoryRecipientCreate, InventoryRecipientOut>({
      url: route("/inventory/recipients"),
      body,
    });
  }

  updateRecipient(id: string, body: InventoryRecipientUpdate) {
    return this.http.put<InventoryRecipientUpdate, InventoryRecipientOut>({
      url: route(`/inventory/recipients/${id}`),
      body,
    });
  }

  deactivateRecipient(id: string) {
    return this.http.post<object, InventoryRecipientOut>({
      url: route(`/inventory/recipients/${id}/deactivate`),
    });
  }

  issue(body: InventoryIssueInput) {
    return this.http.post<InventoryIssueInput, InventoryAssignmentOut>({
      url: route("/inventory/issue"),
      body,
    });
  }

  returnAssignment(id: string, body: InventoryReturnInput) {
    return this.http.post<InventoryReturnInput, InventoryAssignmentOut>({
      url: route(`/inventory/assignments/${id}/return`),
      body,
    });
  }

  addStock(body: InventoryStockAddInput) {
    return this.http.post<InventoryStockAddInput, void>({
      url: route("/inventory/stock"),
      body,
    });
  }

  writeOffStock(body: InventoryStockWriteOffInput) {
    return this.http.post<InventoryStockWriteOffInput, void>({
      url: route("/inventory/write-off"),
      body,
    });
  }

  assignments(params: { itemId?: string; recipientId?: string } = {}) {
    return this.http.get<InventoryAssignmentOut[]>({ url: route("/inventory/assignments", params) });
  }

  transactions(params: Record<string, string | string[] | number | undefined> = {}) {
    return this.http.get<InventoryTransactionOut[]>({ url: route("/inventory/transactions", params) });
  }

  itemTransactions(id: string) {
    return this.http.get<InventoryTransactionOut[]>({ url: route(`/inventory/items/${id}/transactions`) });
  }

  recipientTransactions(id: string) {
    return this.http.get<InventoryTransactionOut[]>({ url: route(`/inventory/recipients/${id}/transactions`) });
  }

  exportURL() {
    return route("/inventory/transactions/export");
  }
}
