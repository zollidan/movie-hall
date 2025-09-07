import { Alert, AlertTitle, AlertDescription } from "./alert";
import { AlertCircle, CheckCircle2 } from "lucide-react";

interface ToastProps {
  type: "success" | "error";
  title: string;
  description?: string;
}

export function Toast({ type, title, description }: ToastProps) {
  return (
    <Alert
      variant={type === "success" ? "default" : "destructive"}
      className="fixed bottom-4 right-4 w-auto"
    >
      {type === "success" ? (
        <CheckCircle2 className="h-4 w-4" />
      ) : (
        <AlertCircle className="h-4 w-4" />
      )}
      <AlertTitle>{title}</AlertTitle>
      {description && <AlertDescription>{description}</AlertDescription>}
    </Alert>
  );
}
