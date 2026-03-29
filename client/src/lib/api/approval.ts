import { request } from "./client";

export interface ApprovalStep {
    ID: number;
    InstanceID: number;
    StepOrder: number;
    ApproverID: number;
    Role: string;
    Status: "PENDING" | "APPROVED" | "REJECTED";
    ActionAt: string | null;
    CreatedAt: string;
    UpdatedAt: string;
}

export interface ApprovalAction {
    ID: number;
    InstanceID: number;
    StepID: number;
    ActorID: number;
    ActionType: "APPROVED" | "REJECTED";
    Comment: string;
    CreatedAt: string;
}

export interface ApprovalInstance {
    ID: number;
    EntityType: string;
    EntityID: number;
    WorkflowID: string;
    Status: "PENDING" | "APPROVED" | "REJECTED";
    CurrentStep: number;
    CreatedBy: number;
    Steps: ApprovalStep[];
    Actions: ApprovalAction[];
    CreatedAt: string;
    UpdatedAt: string;
}

export const approvalApi = {
    /**
     * Get approval instance by entity type and entity ID
     */
    getApprovalStatus: (entityType: string, entityId: number) =>
        request<ApprovalInstance>(`/approval/${entityType}/${entityId}`),

    /**
     * Get approval instance by workflow ID
     */
    getApprovalByWorkflow: (workflowId: string) =>
        request<ApprovalInstance>(`/approval/workflows/${workflowId}`),

    /**
     * Approve the current approval step using workflow ID
     */
    approveStepByWorkflow: (workflowId: string, reason: string) =>
        request<ApprovalInstance>(
            `/approval/workflows/${workflowId}/approve`,
            {
                method: "POST",
                body: JSON.stringify({ comment: reason }),
            }
        ),

    /**
     * Reject the current approval step using workflow ID
     */
    rejectStepByWorkflow: (workflowId: string, reason: string) =>
        request<ApprovalInstance>(
            `/approval/workflows/${workflowId}/reject`,
            {
                method: "POST",
                body: JSON.stringify({ comment: reason }),
            }
        ),
};
