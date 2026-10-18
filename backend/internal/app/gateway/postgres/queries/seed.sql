-- name: SeedDemands :exec
INSERT INTO work_items(id,title,description,kind,priority,status,assignee,reviewer,assignee_id,reviewer_id,reference,event_codes,qa_cycle,created_at,updated_at)
    VALUES
    ('11111111-1111-4111-8111-111111111101','Adequar validação de eventos cadastrais','Aplicar nova regra de validação e documentar cenários esperados.','improvement','high','in_progress','Alex','Marina','55555555-5555-4555-8555-555555555501','55555555-5555-4555-8555-555555555504','DEMO-101',ARRAY['S-2200','S-2205'],0,now()-interval '5 day',now()),
    ('11111111-1111-4111-8111-111111111102','Ajustar importação de arquivo TXT','Garantir compatibilidade com campos opcionais e layouts anteriores.','corrective','critical','in_qa','Rafa','Marina','55555555-5555-4555-8555-555555555502','55555555-5555-4555-8555-555555555504','DEMO-102',ARRAY['S-2500'],1,now()-interval '3 day',now()),
    ('11111111-1111-4111-8111-111111111103','Reescrever serviço de eventos','Isolar regras de domínio e remover acoplamentos legados.','rewrite','medium','backlog','Daniel','Júlia','55555555-5555-4555-8555-555555555503','55555555-5555-4555-8555-555555555505','DEMO-103',ARRAY['S-1202'],0,now()-interval '2 day',now()),
    ('11111111-1111-4111-8111-111111111104','Criar matriz de testes de desligamento','Cobrir casos de afastamento, retificação e desligamento.','testing','high','changes_requested','Alex','Júlia','55555555-5555-4555-8555-555555555501','55555555-5555-4555-8555-555555555505','DEMO-104',ARRAY['S-2299'],1,now()-interval '7 day',now()),
    ('11111111-1111-4111-8111-111111111105','Melhorar relatório de processamento','Incluir filtros de período e resultado por evento.','improvement','low','backlog','Rafa','Marina','55555555-5555-4555-8555-555555555502','55555555-5555-4555-8555-555555555504','DEMO-105',ARRAY['S-2501'],0,now()-interval '1 day',now())
    ON CONFLICT (id) DO NOTHING;

-- name: SeedPackages :exec
INSERT INTO work_packages(id,reference,title,description,initiative,kind,created_at,updated_at) VALUES
 ('22222222-2222-4222-8222-222222222201','OS-DEMO-200','Adequação legal de eventos','Exemplo de OS com trabalho dividido entre profissionais.','Atualização regulatória fictícia','improvement',now()-interval '6 day',now()),
 ('22222222-2222-4222-8222-222222222202','OS-DEMO-210','Melhoria do processamento de lotes','Exemplo de OS executada integralmente por uma pessoa.','Otimização de processos','improvement',now()-interval '4 day',now()) ON CONFLICT(id) DO NOTHING;

-- name: SeedPackageDemands :exec
INSERT INTO work_items(id,title,description,kind,priority,status,assignee,reviewer,assignee_id,reviewer_id,reference,event_codes,qa_cycle,return_count,created_at,updated_at,package_id,product) VALUES
 ('33333333-3333-4333-8333-333333333301','Validação de campos do evento','Item 1 da mesma OS, atribuído a um desenvolvedor.','improvement','high','in_progress','Alex','Marina','55555555-5555-4555-8555-555555555501','55555555-5555-4555-8555-555555555504','OS-DEMO-200',ARRAY['S-2200'],0,0,now()-interval '4 day',now(),'22222222-2222-4222-8222-222222222201','eSocial'),
 ('33333333-3333-4333-8333-333333333302','Ajustar composição do XML','Item 2 da mesma OS, atribuído a outra pessoa.','improvement','high','backlog','Daniel','Marina','55555555-5555-4555-8555-555555555503','55555555-5555-4555-8555-555555555504','OS-DEMO-200',ARRAY['S-2205'],0,0,now()-interval '4 day',now(),'22222222-2222-4222-8222-222222222201','eSocial'),
 ('33333333-3333-4333-8333-333333333303','Refatorar serviço de processamento','Uma única entrega representa todo o trabalho da OS.','improvement','medium','in_progress','Alex','Marina','55555555-5555-4555-8555-555555555501','55555555-5555-4555-8555-555555555504','OS-DEMO-210',ARRAY[]::text[],0,0,now()-interval '3 day',now(),'22222222-2222-4222-8222-222222222202','eSocial'),
 ('33333333-3333-4333-8333-333333333304','Corrigir timeout na integração','Ticket corretivo independente, sem OS agrupadora.','corrective','critical','backlog','Daniel','Marina','55555555-5555-4555-8555-555555555503','55555555-5555-4555-8555-555555555504','BUG-DEMO-500',ARRAY[]::text[],0,0,now()-interval '2 day',now(),NULL,'Integrações') ON CONFLICT(id) DO NOTHING;

-- name: SeedMembers :exec
INSERT INTO users(id,name,email,role,password_hash,created_at) VALUES
 ('55555555-5555-4555-8555-555555555501','Alex','alex@example.test','member',sqlc.arg(password_hash),now()),
 ('55555555-5555-4555-8555-555555555502','Rafa','rafa@example.test','member',sqlc.arg(password_hash),now()),
 ('55555555-5555-4555-8555-555555555503','Daniel','daniel@example.test','member',sqlc.arg(password_hash),now()),
 ('55555555-5555-4555-8555-555555555504','Marina','marina@example.test','member',sqlc.arg(password_hash),now()),
 ('55555555-5555-4555-8555-555555555505','Júlia','julia@example.test','member',sqlc.arg(password_hash),now()) ON CONFLICT DO NOTHING;

-- name: LinkSeedAssignments :exec
WITH unique_names AS (SELECT lower(name) AS name,min(id::text)::uuid AS id FROM users GROUP BY lower(name) HAVING count(*)=1)
 UPDATE work_items w SET assignee_id=COALESCE(w.assignee_id,a.id),reviewer_id=COALESCE(w.reviewer_id,q.id)
 FROM unique_names a,unique_names q WHERE lower(w.assignee)=a.name AND lower(w.reviewer)=q.name
 AND w.id IN ('11111111-1111-4111-8111-111111111101','11111111-1111-4111-8111-111111111102','11111111-1111-4111-8111-111111111103','11111111-1111-4111-8111-111111111104','11111111-1111-4111-8111-111111111105','33333333-3333-4333-8333-333333333301','33333333-3333-4333-8333-333333333302','33333333-3333-4333-8333-333333333303','33333333-3333-4333-8333-333333333304');

-- name: SeedNotifications :exec
INSERT INTO notifications(id,user_id,demand_id,kind,title,message)
 SELECT ('88888888-8888-4888-8888-'||right(u.id::text,12))::uuid,u.id,w.id,'assigned','Sua primeira entrega',
 'Você tem entregas de demonstração para acompanhar. Abra os detalhes e experimente o fluxo de trabalho.'
 FROM users u JOIN LATERAL(SELECT id FROM work_items WHERE assignee_id=u.id OR reviewer_id=u.id ORDER BY created_at,id LIMIT 1) w ON true
 WHERE u.role='member' AND u.id IN ('55555555-5555-4555-8555-555555555501','55555555-5555-4555-8555-555555555502','55555555-5555-4555-8555-555555555503','55555555-5555-4555-8555-555555555504','55555555-5555-4555-8555-555555555505')
 ON CONFLICT(id) DO NOTHING;
