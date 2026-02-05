import { PageHeader, PageShell } from '@/components/ui/page-shell';
import { getActiveProjectId } from '@/lib/server-context';

import { RunCreateForm } from './run-create-form';

export default function NewRunPage() {
  const projectId = getActiveProjectId();

  return (
    <PageShell>
      <PageHeader
        title="Новый RunSpec"
        description="Создайте неизменяемую спецификацию запуска. Все данные фиксируются и используются для детерминированного планирования."
      />
      <RunCreateForm projectId={projectId} />
    </PageShell>
  );
}
