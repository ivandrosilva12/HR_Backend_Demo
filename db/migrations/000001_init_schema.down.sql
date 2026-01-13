-- ============================================================
-- FICHEIRO ÚNICO: Limpeza + Esquema atualizado (fusão consolidada)
-- Mantém integridade referencial, histórico de posições (worker_histories),
-- hierarquia de departamentos, capacidade de posições e funções de pesquisa.
-- Seguro para reexecução (DROP IF EXISTS / IF NOT EXISTS).
-- ============================================================

-- ========= PRE-REQUISITOS ===================================
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- ========= TEARDOWN / LIMPEZA ===============================

-- 1) Views (as dependências mais frágeis primeiro)
DROP VIEW IF EXISTS positions_with_occupancy;
DROP VIEW IF EXISTS departments_tree;
DROP VIEW IF EXISTS view_positions_search;
DROP VIEW IF EXISTS view_municipalities_search;

-- 2) Triggers conhecidos e específicos (precisam das tabelas existentes)
-- Employees
DO $$
BEGIN
  IF to_regclass('public.employees') IS NOT NULL THEN
    EXECUTE 'DROP TRIGGER IF EXISTS trg_worker_history_after_insert ON employees';
    EXECUTE 'DROP TRIGGER IF EXISTS trg_worker_history_after_update ON employees';
    EXECUTE 'DROP TRIGGER IF EXISTS trg_enforce_position_rules ON employees';
  END IF;
END $$;

-- Departments
DO $$
BEGIN
  IF to_regclass('public.departments') IS NOT NULL THEN
    EXECUTE 'DROP TRIGGER IF EXISTS trg_department_cycle ON departments';
    EXECUTE 'DROP TRIGGER IF EXISTS trg_prevent_delete_parent_department ON departments';
  END IF;
END $$;

-- 3) Triggers utilitárias de updated_at (em loop) – ainda com tabelas presentes
DO $$
DECLARE
    tbl TEXT;
BEGIN
    FOR tbl IN SELECT unnest(ARRAY[
        'areas_estudo', 'provinces', 'departments', 'positions',
        'municipalities', 'districts', 'employees', 'documents',
        'employee_statuses', 'dependents', 'education_histories',
        'work_histories', 'supervisor_histories', 'worker_histories'
    ])
    LOOP
        IF to_regclass('public.'||tbl) IS NOT NULL THEN
          EXECUTE format('DROP TRIGGER IF EXISTS trg_set_updated_at_%I ON %I;', tbl, tbl);
        END IF;
    END LOOP;
END;
$$;

-- 4) Funções (buscas e utilitárias) – podem ser recriadas depois
-- Buscas antigas/atuais
DROP FUNCTION IF EXISTS search_supervisor_histories(uuid, date, date, int, int);
DROP FUNCTION IF EXISTS search_work_histories(uuid, text, date, date, int, int);
DROP FUNCTION IF EXISTS search_education_histories(uuid, text, date, date, int, int);
DROP FUNCTION IF EXISTS search_dependents(text, uuid, int, int);
DROP FUNCTION IF EXISTS search_employee_statuses(uuid, text, int, int);
DROP FUNCTION IF EXISTS search_documents(text, text, text, int, int);
DROP FUNCTION IF EXISTS search_employees(text, text, integer, integer);
DROP FUNCTION IF EXISTS search_districts(text, text, int, int);
DROP FUNCTION IF EXISTS search_areas_estudo(text, int, int);
DROP FUNCTION IF EXISTS search_provinces(text, int, int);
DROP FUNCTION IF EXISTS search_positions(text, text, int, int);
DROP FUNCTION IF EXISTS search_departments(text, int, int);
DROP FUNCTION IF EXISTS search_municipalities(text, text, int, int);
DROP FUNCTION IF EXISTS search_worker_histories(uuid, uuid, text, date, date, int, int);
DROP FUNCTION IF EXISTS list_employees_by_department(uuid, boolean, boolean, int, int);

-- Utilitárias/Triggers
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP FUNCTION IF EXISTS prevent_department_cycle();
DROP FUNCTION IF EXISTS prevent_delete_parent_department();
DROP FUNCTION IF EXISTS enforce_position_rules();
DROP FUNCTION IF EXISTS worker_history_on_insert();
DROP FUNCTION IF EXISTS worker_history_on_position_change();

-- 5) Índices (opcional; cairiam com as tabelas, mas garantimos limpeza)
DROP INDEX IF EXISTS unique_district_per_municipality;
DROP INDEX IF EXISTS unique_municipality_per_province;

DROP INDEX IF EXISTS idx_positions_department;
DROP INDEX IF EXISTS idx_departments_parent;

DROP INDEX IF EXISTS idx_work_employee;
DROP INDEX IF EXISTS idx_worker_employee;
DROP INDEX IF EXISTS idx_worker_position;
DROP INDEX IF EXISTS idx_worker_active;

DROP INDEX IF EXISTS idx_employee_status_employee;
DROP INDEX IF EXISTS idx_employee_status_current;
DROP INDEX IF EXISTS idx_education_employee;
DROP INDEX IF EXISTS idx_dependents_employee;

DROP INDEX IF EXISTS idx_documents_owner;
DROP INDEX IF EXISTS idx_documents_type;
DROP INDEX IF EXISTS idx_documents_ownerid_uploadedat;
DROP INDEX IF EXISTS idx_documents_owner_type_date;
DROP INDEX IF EXISTS idx_documents_object_key;

DROP INDEX IF EXISTS idx_employees_bi;
DROP INDEX IF EXISTS idx_employees_email;
DROP INDEX IF EXISTS idx_employees_department_id;
DROP INDEX IF EXISTS idx_employees_position_id;
DROP INDEX IF EXISTS idx_employees_district_id;

-- 6) Tabelas (ordem inversa à criação típica)
DROP TABLE IF EXISTS supervisor_histories;
DROP TABLE IF EXISTS worker_histories;
DROP TABLE IF EXISTS work_histories;
DROP TABLE IF EXISTS education_histories;
DROP TABLE IF EXISTS dependents;
DROP TABLE IF EXISTS employee_statuses;
DROP TABLE IF EXISTS documents;
DROP TABLE IF EXISTS employees;
DROP TABLE IF EXISTS districts;
DROP TABLE IF EXISTS municipalities;
DROP TABLE IF EXISTS positions;
DROP TABLE IF EXISTS departments;
DROP TABLE IF EXISTS provinces;
DROP TABLE IF EXISTS areas_estudo;

