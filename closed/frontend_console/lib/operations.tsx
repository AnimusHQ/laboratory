'use client';

import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react';

export type OperationStatus = 'pending' | 'succeeded' | 'failed';

export type ConsoleOperation = {
  id: string;
  label: string;
  status: OperationStatus;
  createdAt: string;
  updatedAt?: string;
  details?: string;
};

type OperationsContextValue = {
  operations: ConsoleOperation[];
  addOperation: (operation: ConsoleOperation) => void;
  updateOperation: (id: string, status: OperationStatus, details?: string) => void;
  clearCompleted: () => void;
};

const STORAGE_KEY = 'animus_console_operations_v1';

const OperationsContext = createContext<OperationsContextValue | null>(null);

const loadOperations = (): ConsoleOperation[] => {
  if (typeof window === 'undefined') {
    return [];
  }
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    if (!raw) {
      return [];
    }
    const parsed = JSON.parse(raw) as ConsoleOperation[];
    return Array.isArray(parsed) ? parsed : [];
  } catch {
    return [];
  }
};

const persistOperations = (operations: ConsoleOperation[]) => {
  if (typeof window === 'undefined') {
    return;
  }
  try {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(operations));
  } catch {
    // ignore persistence failures
  }
};

export function OperationsProvider({ children }: { children: React.ReactNode }) {
  const [operations, setOperations] = useState<ConsoleOperation[]>([]);

  useEffect(() => {
    setOperations(loadOperations());
  }, []);

  useEffect(() => {
    persistOperations(operations);
  }, [operations]);

  const addOperation = useCallback((operation: ConsoleOperation) => {
    setOperations((prev) => {
      const next = [operation, ...prev].slice(0, 50);
      return next;
    });
  }, []);

  const updateOperation = useCallback((id: string, status: OperationStatus, details?: string) => {
    setOperations((prev) =>
      prev.map((operation) =>
        operation.id === id
          ? {
              ...operation,
              status,
              details: details ?? operation.details,
              updatedAt: new Date().toISOString(),
            }
          : operation,
      ),
    );
  }, []);

  const clearCompleted = useCallback(() => {
    setOperations((prev) => prev.filter((op) => op.status === 'pending'));
  }, []);

  const value = useMemo(
    () => ({
      operations,
      addOperation,
      updateOperation,
      clearCompleted,
    }),
    [operations, addOperation, updateOperation, clearCompleted],
  );

  return <OperationsContext.Provider value={value}>{children}</OperationsContext.Provider>;
}

export function useOperations() {
  const ctx = useContext(OperationsContext);
  if (!ctx) {
    throw new Error('useOperations must be used within OperationsProvider');
  }
  return ctx;
}
